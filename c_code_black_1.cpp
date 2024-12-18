cellularif_cmd_exec_ret_t cellularif_cmd_shutdown_process(cellularif_cmd_t* cmd); static void cellularif_status_iccid_diff_process(bool first_time,
                                                 cellularif_status_t* new_status)
{
    if (s_cellularif_conf.enable_bind_sim && s_cellularif_conf.bind_iccid == NULL &&
        new_status->iccid != NULL) {
        log_debug("networkd", "", "[info] SIM lazy binding, bind with: %s\n", new_status->iccid);
        s_cellularif_conf.bind_iccid = strdup(new_status->iccid);
        write_iccid(s_cellularif_conf.bind_iccid);
    }

    if (s_cellularif_conf.enable_bind_sim && s_cellularif_conf.bind_iccid != NULL &&
        new_status->iccid != NULL) {
        if (strncmp(s_cellularif_conf.bind_iccid, new_status->iccid,
                    strlen(s_cellularif_conf.bind_iccid) != 0)) {
            log_debug("networkd", "", "[info] iccid %s not equal to %s(bound)\n",
                      s_cellularif_conf.bind_iccid, new_status->iccid);
            cellularif_cmd_t cmd = {.cmd = CELLULARIF_CMD_HANG, .arg.hang = 0};
            cellularif_cmd_shutdown_process(&cmd);
        }
    }
}