bool cellularif_create(const cellularif_conf_t* conf, struct interface* ifp)
{
    if (ifp == NULL) {
        log_debug("networkd", "", "[error] NULL ifp\n");
        return false;
    }

    if (s_cellularifp != NULL) {
        log_debug("networkd", "", "[error] cellularif already created\n");
        return false;
    }

    if (!cellularif_conf_check(conf)) {
        log_debug("networkd", "", "[error] invalid conf\n");
        return false;
    }

    cellularif_cmd_queue_init(&s_cellularif_daemon_cmd);

    cellularif_conf_copy(&s_cellularif_conf, conf);
    s_cellularif_conf.bind_iccid = read_iccid();
    if (s_cellularif_conf.bind_iccid != NULL) {
        s_cellularif_conf.enable_bind_sim = true;
    }

    if (!launch_cellularif_daemon_threads()) {
        log_debug("networkd", "", "[error] failed to launch cellularif daemon threads\n");
        cellularif_conf_clear(&s_cellularif_conf);
        return false;
    }

    s_cellularifp = ifp;

    cellularif_shutdown();

    return true;
}