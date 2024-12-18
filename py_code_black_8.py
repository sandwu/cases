from django.db.models.signals import post_save, post_delete
@classmethod
def register_callbacks(cls):
        from execution.logstash.callbacks import handle_strategy_modify, handle_rule_modified, \
            handle_custom_rule_modified
        from webapi.models.strategy import StrategyModel, RuleModel, RuleCustomizeModel
        monitor_signals = [post_save, post_delete]
        for monitor_signal in monitor_signals:
            monitor_signal.connect(handle_strategy_modify, sender=StrategyModel,
                                   dispatch_uid="handle_strategy_modify {}".format(id(monitor_signal)))
            monitor_signal.connect(handle_rule_modified, sender=RuleModel,
                                   dispatch_uid="handle_rule_modified {}".format(id(monitor_signal)))
            monitor_signal.connect(handle_custom_rule_modified, sender=RuleCustomizeModel,
                                   dispatch_uid="handle_custom_rule_modified {}".format(id(monitor_signal)))
        logger.info("django model signals registered")