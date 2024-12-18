/*
* 判断是绕流会话时,初始化扩展指针, 如果指针已经初始化, 不重复初始化
*/
int sase_rt_sess_ext_hb_init(sase_rt_session_t* sr_sess, ssf_data_context_t *ctx)
{
    int need_init = 0;
    if (sr_sess == NULL || sr_sess->mode != ROUTE_MODE_IOC || sr_sess->ext == NULL) return -1;

    dp_plugins_session_lock(ctx);
    if (sr_sess->ext->hb == NULL) {
        need_init = 1;
        sr_sess->ext->hb = dp_plugins_malloc(sizeof(sase_rt_heartbeat_t));
    }
    dp_plugins_wmb();
    dp_plugins_session_lock(ctx);
    if (sr_sess->ext->hb && need_init) {
        memset(sr_sess->ext->hb, 0, sizeof(sase_rt_heartbeat_t));
        sr_sess->ext->hb->last_hb_secs = dp_plugins_time_secs();
    }
    
    return sr_sess->ext->hb? 0: -1;
}