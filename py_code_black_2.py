from datetime import datetime
from governance.enums import MetricEntityType, MetricFieldType
def _process_metric_by_strategy_id(
        daily_metric_models, metric_list_by_strategy_id, strategy_id,
        strategy_model
):
    from webapi.models import MetricModel, StrategyDailyMetricStatModel
    strategy_daily_metric_map = {}
    for metric_model in metric_list_by_strategy_id:
        metric_model: MetricModel = metric_model
        try:
            metric_entity_type: MetricEntityType = \
                MetricEntityType(metric_model.entity)
            metric_field_type: MetricFieldType = \
                MetricFieldType(metric_model.metric_field)
        except ValueError:
            continue
        if metric_entity_type != MetricEntityType.STRATEGY:
            # 暂不考虑其他种类的度量
            continue
        # 统计EPS
        if metric_field_type not in strategy_daily_metric_map:
            strategy_daily_metric_map[metric_field_type] = MetricRateStat()
        strategy_daily_metric_map[metric_field_type].add(metric_model)
        # 示例 2022-05-09 11:03:54
        end_datetime = datetime.fromtimestamp(metric_model.end)
        # 示例 2022-05-09
        end_date = end_datetime.date()
        # 更新接入时间
        # fix: 每次使用最新时间作为同步时间，避免时间调整导致时间停止更新
        strategy_model.sync_on = end_datetime
        # 获取相关日期的度量
        daily_model_key = f'{strategy_id}-{end_date}'
        if daily_model_key in daily_metric_models:
            daily_metric_model: StrategyDailyMetricStatModel = \
                daily_metric_models[daily_model_key]
        else:
            daily_metric_model: StrategyDailyMetricStatModel = \
                StrategyDailyMetricStatModel.objects.filter(
                    date=end_date, strategy_id=strategy_id).first()
            if daily_metric_model is None:
                daily_metric_model: StrategyDailyMetricStatModel = \
                    StrategyDailyMetricStatModel()
                daily_metric_model.strategy_id = strategy_id
                daily_metric_model.date = end_date
            daily_metric_models[daily_model_key] = daily_metric_model

        if metric_field_type in _process_mapping:
            _process_mapping[metric_field_type](
                strategy_model,
                metric_model.count,
                metric_model.volume,
                _judge_today(metric_model.end),
                metric_model.end - metric_model.start,
                daily_metric_model
            )
    _process_eps(strategy_daily_metric_map, strategy_model)