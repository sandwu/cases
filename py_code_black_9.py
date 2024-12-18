@classmethod
def classify_tenant_rules_from_etcd(cls, rules):
        tenant_rule_dict = {}
        for customize_rule in rules:
            rule_name = customize_rule[1].key.decode('utf-8')
            if rule_name.startswith(cls.ETCD_PREFIX):
                rule_name = rule_name[len(cls.ETCD_PREFIX):]
            else:
                logger.error("bad key: {}".format(rule_name))
                continue
            tenant, rule_id = rule_name.split("-")[1:]
            rule_id = int(rule_id)
            rule_content = customize_rule[0].decode('utf-8')
            if tenant in tenant_rule_dict:
                tenant_rule_dict[tenant][rule_id] = rule_content
            else:
                tenant_rule_dict[tenant] = {rule_id: rule_content}
        return tenant_rule_dict