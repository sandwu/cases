int parse_pci_info( char *buf, int size, sf_uint8_t * is_dpdk_support, sf_uint16_t portid)
{
    sf_int32_t n = 0;
    memset(buf, 0, size);

    sf_device_nic_info_t* nic_info = sf_get_phy_nic_info(portid);
    struct sf_pci_device* pci_dev = sf_get_phy_nic_pci_info(portid);
    if (pci_dev == NULL) {
        log_debug("dp driver", "", "[error]portid:%u pci-info is error\n", portid);
        return -1;
    }

	// 芯启源网卡使用PCI类型字符串
    if (sf_is_smartnic_nfp_by_driver((char *)nic_info->driver_name)) {
        const struct sf_pci_addr *loc = &(pci_dev->addr);
        *is_dpdk_support = 1;
        sf_int16_t vf_num = sf_smart_nic_nfp_get_vfnum(portid);

        //启动参数中添加VF信息
        n = snprintf(buf, size, SMART_PCI_PRI_FMT,
                loc->domain, loc->bus, loc->devid, loc->function, vf_num - 1);
        log_debug("dp driver", "", "[info]:nic_info->driver_name %s pci buf: %s", (char *)nic_info->driver_name, buf);
        if ((n < 0) || (n >= size)) {
            log_debug("dp driver", "", "[error]:portid:%u snprintf fail, n=%d", portid, n);
            return -1;
        }
    } else {
        if (!sf_dev_nic_is_dpdk_support(portid)) {
            if (sf_dev_nic_is_cellular_port(portid)) {
                log_debug("dp driver", "", "[info]: portid: %u is cellular port", portid);
                n = snprintf(buf, size, SF_GENERIC_DRIVER_CELLULAR_NAME_FMT, cellular_if_idx,
                             cellular_if_idx, sf_get_generic_port_queue_num(portid));
                if (n < 0 || n >= size) {
                    log_debug("dp driver", "",
                              "[error]: failed to build cellular port args, portid: %u, n: %d",
                              portid, n);
                    return -1;
                }
                cellular_if_adapt_port_id = portid;
                cellular_if_idx++;
            } else {
                log_debug("dp driver", "", "[info]: portid: %u is generic port", portid);
                n = snprintf(buf, size, SF_GENERIC_DRIVER_NAME_FMT, portid - 1, portid - 1,
                             sf_get_generic_port_queue_num(portid));
                if ((n < 0) || (n >= size)) {
                    log_debug("dp driver", "",
                              "[error]:not dpdk support portid:%u snprintf fail, n=%d", portid,
                              n);
                    return -1;
                }
            }
        } else if (sf_dev_nic_is_switch_phy_port(portid)) {
            const struct sf_pci_addr* loc = &(pci_dev->addr);
            n = snprintf(buf, size, SF_SWITCH_PHY_PORT_DRIVER_NAME_FMT, portid - 1, loc->domain,
                         loc->bus, loc->devid, loc->function);
            if ((n < 0) || (n >= size)) {
                log_debug("dp driver", "", "[error]:switch phy portid:%u snprintf fail, n=%d",
                          portid, n);
                return -1;
            }
        } else {
            //得到pci的字串
            const struct sf_pci_addr* loc = &(pci_dev->addr);
            *is_dpdk_support = 1;
            n = snprintf(buf, size, PCI_PRI_FMT, loc->domain, loc->bus, loc->devid,
                         loc->function);
            if ((n < 0) || (n >= size)) {
                log_debug("dp driver", "", "[error]:portid:%u snprintf fail, n=%d", portid, n);
                return -1;
            }
        }
    }

    return 0;
}