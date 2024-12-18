int is_dp_fw_running(void)
{
    int status = 0;
    int exit_status = 0;
    int out_len = 128;
    char out[128] = {0};
    imdtsk_rcfg_t rcfg = {0};
    rcfg.timeout = 10;
    rcfg.need_status = 1;
    rcfg.need_out = 1;

    if (imdtsk_clt_run(CMD_PIDOF_DP_FW, &rcfg, &status, out, &out_len) < 0)
    return 1;

    if (WIFEXITED(status)) {
    exit_status = WEXITSTATUS(status);
    // 正常退出说明在运行
    if (exit_status == 0)
        return 1;
    else
        // 不正常退出说明不存在
        return 0;
    }
    // 异常退出认为正在运行
    return 1;
}