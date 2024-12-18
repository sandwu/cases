from typing import List, Union
from dumper.pulsar_input import CustomMessage
from execution.logstash.config import PULSAR_CUSTOM_MSG_DEL_PATH
from governance.strategy import Strategy
def _reload_pipeline_cfg(
            self,
            strategies: List[Strategy],
            delete: bool = False,
            msg_list: List[Union[str, CustomMessage]] = None
    ):
        """刷新pipeline的缓存

        pipeline缓存：
        - `self._passive_monitor_pipeline_by_id`
        - `self._active_monitor_pipeline_by_id`

        填充缓存：
        - 创建 input 管道配置文件
        - 创建 receiver 管道配置文件

        使用缓存：
        - 在 `_reload_monitor_pipeline_config` 中创建 pipeline.yml 配置文件时
        """
        last_passive_pipelines = dict(self._passive_monitor_pipeline_by_id)
        last_active_pipelines = dict(self._active_monitor_pipeline_by_id)
        self._passive_monitor_pipeline_by_id = dict()
        self._active_monitor_pipeline_by_id = dict()
        self._create_pipeline_cfg(strategies)

        if not delete:
            return

        removed_paths = []
        removed_paths.extend(self._get_deleted_pipeline_cache(
            last_passive_pipelines,
            self._passive_monitor_pipeline_by_id))
        removed_paths.extend(self._get_deleted_pipeline_cache(
            last_active_pipelines,
            self._active_monitor_pipeline_by_id))

        logger.debug("removed pipeline files: %s", removed_paths)

        if removed_paths and isinstance(msg_list, list):
            msg_list.append(CustomMessage(
                msg_type=PULSAR_CUSTOM_MSG_DEL_PATH,
                msg_detail=list(removed_paths)))