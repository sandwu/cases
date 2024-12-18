static bool cellularif_cmd_dispatch(const cellularif_cmd_t* cmd)
{
    // log_debug("networkd", "", "[debug] dispatch %s\n", str_cellularif_cmd_type(cmd->cmd));

    if (!cellularif_cmd_check(cmd)) {
        log_debug("networkd", "", "[error] cmd is invalid\n");
        return false;
    }

    int ret = pthread_mutex_lock(&s_cellularif_daemon_mtx);
    if (ret != 0) {
        log_debug("networkd", "", "[error] failed to lock cellular daemon mtx %s(%d)\n",
                  strerror(ret), ret);
        return false;
    }

    bool stat = cellularif_cmd_queue_append(&s_cellularif_daemon_cmd, cmd);

    pthread_cond_signal(&s_cellularif_daemon_cond);
    pthread_mutex_unlock(&s_cellularif_daemon_mtx);

    return stat;
}