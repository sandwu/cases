void GlobalCfg::GetDnsData(const char *section, unsigned char flag,vector<global_ctl_dns_st> &dnsdata)
{
    int nInnerCount = iniInner.GetLong(section, CONFIG_ITEM_DOMAINCOUNT);

    if(flag == WHITE_FLAG) {
    nInnerCount = get_min(nInnerCount, m_white_domain_limit, DEFAULT_GLOBAL_CTL_WHITE_DOMAIN_LIMIT);
    }

    char buf[256] = { '\\0' };
    char key[32] = { '\\0' };
    int index = 0;
    for (int i = 0; i < nInnerCount; ++i) {
    (void)snprintf(key, sizeof(key), CONFIG_ITEM_DNSCHK, i);
    if (0 == iniInner.GetLong(section, key, 1)) {
        continue;
    }

    memset(buf, 0, sizeof(buf));
    (void)snprintf(key, sizeof(buf), CONFIG_ITEM_DNS, i);
    iniInner.GetString(section, key, buf, sizeof(buf));
    if (buf[0]) {
        ac_to_lower(buf);
        global_ctl_dns_st dns;
        dns.hosts_info.dnsCRC = crc32((const unsigned char *)buf, strlen(buf));
        dns.hosts_info.flag = flag;
        dns.hosts_info.index = index;
        strcpy_n(dns.dns, MAX_GLOBAL_CTL_DNS_LEN, buf);
        dnsdata.push_back(dns);
        index++;
    } 
    }
}