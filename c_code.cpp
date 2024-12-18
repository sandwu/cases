void dp_overload_protection_enable_show(struct idl_param_st *p, struct idl_buf_st *output, void *userdat)
{
    struct dp_overload_context_t *ctx = dp_overload_ctx();
    assert(ol);
    if (output != NULL) {
        idl_buf_printf(output, DP_METRICS_STR_MAX, "%s", ctx->cfg.enable ? "enable" : "disable");
    }
}

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

cellularif_auth_protocol_t cellularif_auth_protocol_from_str(const char* str)
{
    if (str == NULL)
        return CELLULARIF_AUTH_PROTOCOL_NONE;

    if (strncmp(str, "AUTO", strlen(str)) == 0)
        return CELLULARIF_AUTH_PROTOCOL_AUTO;
    if (strncmp(str, "PAP", strlen(str)) == 0)
        return CELLULARIF_AUTH_PROTOCOL_PAP;
    if (strncmp(str, "CHAP", strlen(str)) == 0)
        return CELLULARIF_AUTH_PROTOCOL_CHAP;

    return CELLULARIF_AUTH_PROTOCOL_NONE;
}

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

int dp_overload_latency_is_overload(struct dp_overload_context_t *ctx,
    struct latency_distribute_stats *stats)
{
    int result = 0;

    sf_uint64_t max_us = dp_latency_stats_get_max_us(stats);
    sf_uint64_t avg_us = dp_latency_stats_get_avg_us(stats);
    sf_uint64_t exceed_max_cnt = dp_latency_stats_exceed_threshold_count(stats, max_us);
    sf_uint64_t stats_cnt = dp_latency_stats_exceed_threshold_count(stats, 0);

    if (ctx->cfg.latency_max_us_thr > 0) {
        if (max_us >= ctx->cfg.latency_max_us_thr
            && exceed_max_cnt >= ctx->cfg.latency_max_us_thr) {
            (void)dp_ol_reason_append(ctx, "sample max latency %lu(times %lu) reached threshold %lu",
                max_us, exceed_max_cnt, ctx->cfg.latency_max_us_thr);
            result = 1;
        } else {
            (void)dp_ol_reason_append(ctx, "sample max latency %lu(times %lu) unreached threshold %lu",
                max_us, exceed_max_cnt, ctx->cfg.latency_max_us_thr);
        }
    }
    
    if (ctx->cfg.latency_avg_us_thr > 0) {
        (void)dp_ol_reason_append(ctx, ",");
        if (avg_us >= ctx->cfg.latency_avg_us_thr) {
            (void)dp_ol_reason_append(ctx, "sample avg latency %lu(stats cnt %lu) reached threshold %lu",
                avg_us, stats_cnt, ctx->cfg.latency_avg_us_thr);
            result = 1;
        } else {
            (void)dp_ol_reason_append(ctx, "sample avg latency %lu(stats cnt %lu) unreached threshold %lu",
                avg_us, stats_cnt, ctx->cfg.latency_avg_us_thr);
        }
    }
    return result;
}

cellularif_cmd_exec_ret_t cellularif_cmd_shutdown_process(cellularif_cmd_t* cmd); static void cellularif_status_iccid_change_process(bool first_time,
                                                 cellularif_status_t* new_status)
{
    char* prev_iccid = s_cellularif_status.iccid;

    s_cellularif_status.iccid = new_status->iccid;
    new_status->iccid = NULL;

    if (s_cellularif_conf.enable_bind_sim && s_cellularif_conf.bind_iccid == NULL &&
        s_cellularif_status.iccid != NULL) {
        log_debug("networkd", "", "[info] SIM lazy binding, bind with: %s\n",
                  s_cellularif_status.iccid);
        s_cellularif_conf.bind_iccid = strdup(s_cellularif_status.iccid);
        write_iccid(s_cellularif_conf.bind_iccid);
    }

    if (s_cellularif_conf.enable_bind_sim && s_cellularif_conf.bind_iccid != NULL && s_cellularif_status.iccid != NULL) {
        if (strncmp(s_cellularif_conf.bind_iccid, s_cellularif_status.iccid,
                    strlen(s_cellularif_conf.bind_iccid) != 0)) {
            cellularif_cmd_t cmd = {.cmd = CELLULARIF_CMD_HANG, .arg.hang = 0};
            cellularif_cmd_shutdown_process(&cmd);
        }
    }

    free(prev_iccid);
}

int dp_overload_level_update(struct dp_overload_context_t *ctx, 
    sf_uint32_t overload_level)
{
    sf_uint32_t old_overload_level = ctx->status.overload_level;
    if (overload_level >= VNET_OL_LEVEL_MAX) {
        log_debug("overload", "", "[error]overload_level %u invalid.\n", overload_level);
        return -1;
    }
    struct dp_perf_metrics_t *detect_start = dp_perf_metrics_store_front(ctx->store);
    assert(detect_start);
    struct dp_perf_metrics_t *last_metrics = dp_perf_metrics_store_last(ctx->store);
    assert(last_metrics);
    struct dp_perf_metrics_t *now_metrics = dp_perf_metrics_store_now(ctx->store);
    assert(now_metrics);
    struct dp_overload_level_status_t *old_status = &ctx->status.levels[old_overload_level];

    dp_overload_level_set(ctx, overload_level, 0);
    struct dp_overload_level_status_t *new_status = dp_overload_current_level_status(ctx);
    assert(new_status);
    log_debug("performance", "overload", "[warn]system perf metrics %s\n",
        dp_perf_metrics_format(now_metrics, detect_start));
    //过载等级加强
    if (old_overload_level < overload_level) {
        memcpy(&old_status->detect_start_metrics, detect_start, sizeof(old_status->detect_start_metrics));
        memcpy(&old_status->sample_last_metrics, last_metrics, sizeof(old_status->sample_last_metrics));
        memcpy(&old_status->detect_end_metrics, now_metrics, sizeof(old_status->detect_end_metrics));

        //跳级场景，前面copy对应信息
        for (int i = old_overload_level + 1; i < overload_level; i++) {
            if (i >= sizeof(ctx->status.levels) / sizeof(ctx->status.levels[0])) {
                continue;
            }
            memcpy(&ctx->status.levels[i], old_status, sizeof(ctx->status.levels[i]));
        }
        if (VNET_OL_LEVEL1 == overload_level) {
            dp_overload_bypass_percent_set(new_status, tcp_proxy_bypass_percent(ctx, &now_metrics->tcp_proxy));
        } else if (VNET_OL_LEVEL2 == overload_level) {
            dp_overload_bypass_percent_set(new_status, PERCENT_30);
        } else if (VNET_OL_LEVEL3 == overload_level) {
            dp_overload_bypass_percent_set(new_status, PERCENT_100);
        }

        //TODO:支持中英文翻译和字段标准化映射
        log_warn(_T("performance"), _T("overload"), _F("system overloaded, level {1d} -> {2d}, reason {3}", 
            old_overload_level, ctx->status.overload_level, ctx->status.reason),
            _F("please check the performance or configuration"));
        log_debug("performance", "overload", "[warn]system overloaded, level %u -> %u, reason %s\n", 
            old_overload_level, ctx->status.overload_level, ctx->status.reason);
    } else {
        //不需要清理old_status 后面过载等级加强时 刷新即可
        log_warn(_T("performance"), _T("overload"), _F("system cancel overloaded, level {1d} -> {2d}, reason {3}",
            old_overload_level, ctx->status.overload_level, ctx->status.reason),
            _F("please check the performance or configuration"));
        log_debug("performance", "overload", "[warn]system cancel overloaded, level %u -> %u, reason %s\n",
            old_overload_level, ctx->status.overload_level, ctx->status.reason);
    }
    return 0;
}

SF_CORE_INIT_REGISTER(dp_flow_elephant_config_init, "elephant_flow"); static inline int dp_flow_elephant_detect_rate_check(
    sf_uint64_t pkt_bytes, void *session, sf_uint8_t is_ipv4, sf_uint8_t is_uplink)
{
    sf_uint16_t Mbps;
    sf_uint16_t kpps;
    sf_uint32_t session_livetime;
    sf_uint8_t is_keepalive;

    if (is_ipv4) {
        sf_ipv4_session_t *ipv4_session = (sf_ipv4_session_t *)session;
        if (is_uplink){
            Mbps = ipv4_session->flow0.Mbps;
            kpps = ipv4_session->flow0.kpps;
        } else {
            Mbps = ipv4_session->flow1.Mbps;
            kpps = ipv4_session->flow1.kpps;
        }
        session_livetime = (sf_uint32_t)sf_time_s() - ipv4_session->create_time;
        is_keepalive = sf_session_is_set_alive_flag(ipv4_session) ? 1 : 0;
    } else {
        sf_ipv6_session_t *ipv6_session = (sf_ipv6_session_t *)session;
        if (is_uplink){
            Mbps = ipv6_session->flow0.Mbps;
            kpps = ipv6_session->flow0.kpps;
        } else {
            Mbps = ipv6_session->flow1.Mbps;
            kpps = ipv6_session->flow1.kpps;
        }
        session_livetime = (sf_uint32_t)sf_time_s() - ipv6_session->create_time;
        is_keepalive = sf_ipv6_session_is_set_alive_flag(ipv6_session) ? 1 : 0;
    }

    if (((pkt_bytes >= g_elephant_flow_detect_main.total_rate) &&
        (session_livetime >= g_elephant_flow_detect_main.session_duration) &&
        !is_keepalive) ||
        // Session Duration exceeds x s & Total Bits exceeds x Mb (excluding keepalive session)
        (Mbps >= g_elephant_flow_detect_main.instantaneous_rate) ||
        // OR Instantaneous Rate exceeds x Mbps
        (kpps >= g_elephant_flow_detect_main.packet_per_second)) {
        // OR Packet Per Second exceeds x Kpps
        return 1;
    }

    return 0;
}

static int do_upgrade(const char *tenant, const char *db, const char *path)
{
	char cfgpath[PATH_MAX] = {0};
	uint32_t cnt = cc_schemas_count();
	uint32_t i;
	int res;

	cc_cdb_st *cdb = cc_cdb_get(tenant, db);
	if (!cdb) {
		cc_set_last_error("%s", _T("CDB NOT FOUND"));
		log_debug("cfg-center", "upgrader", "Get cdb for tenant(%s) db(%s) failed: %s", tenant, db, cc_last_error());
		record_error(_T("CDB NOT FOUND"), NULL, NULL);
		return -1;
	}

	for (i = 1; i < cnt; i++) { //0号Schema为NULL，不被使用
		const char *cfg = cc_schema_get_cfgname_by_id(i);
		assert(cfg);

		str_snprintf(cfgpath, PATH_MAX, "%s/%s.json", path, cfg);

		/* 待升级的配置不存在，那就不升级此配置了，可能是由于新版本增加的配置，老配置中没有 */
		if (access(cfgpath, F_OK) != 0) {
			debug("The path(%s) is not exists.", cfgpath);
			/* 删掉老文件 */
			str_snprintf(cfgpath, PATH_MAX, "%s/%s/%s/%s.base", cc_cache_db_path(), tenant, db, cfg);
			unlink(cfgpath);
			str_snprintf(cfgpath, PATH_MAX, "%s/%s/%s/%s.actions", cc_cache_db_path(), tenant, db, cfg);
			unlink(cfgpath);
			str_snprintf(cfgpath, PATH_MAX, "%s/%s/%s/%s.digest", cc_cache_db_path(), tenant, db, cfg);
			unlink(cfgpath);

			/* 如果没有指定要平滑的配置，则使用默认配置 */
			idl_object_st *def = cc_cache_get_default_object(cfg, 1);
			assert(def);

			if ((res = cc_cdb_config_set(cdb, cfg, def))) {
				log_debug("cfg-center", "upgrader", "Reset tenant(%s) db(%s) cfg(%s) failed: %s", tenant, db, cfg, cc_last_error());
				idl_object_release(def);
				record_error(cc_last_error(), _T(cfg), NULL);
				return cc_errno_make(res);
			}

			log_debug("cfg-center", "upgrader", "Reset tenant(%s) db(%s) cfg(%s) success", tenant, db, cfg);
			idl_object_release(def);
			continue;
		}

		/* 配置解析成json失败，终止导入 */
		idl_json_st *oldjs = idl_file_parse_json_file(cfgpath);
		if (!oldjs) {
			debug("Failed to get json for config %s: %s", cfg, cc_last_error());
			record_error(_F("{1} format error", "JSON"), _T(cfg), NULL);
			return -1;
		}

		char *old_cfgstr = idl_json_pack(oldjs, NULL);
		idl_json_release(oldjs);

		char *new_cfgstr = NULL;

		if ((res = cc_upgrader_call(tenant, db, cfg, old_cfgstr, &new_cfgstr))) {
			debug("Call upgrade scripts for config %s failed: %s", cfg, cc_last_error());
			free(old_cfgstr);
			record_error(_T("Execute convert exception"), _T(cfg), NULL);
			return cc_errno_make(res);
		}

		/* 解析成json，失败得报错 */
		idl_json_st *newjs = idl_file_parse_json(new_cfgstr, strlen(new_cfgstr));
		if (!newjs) {
			debug("Parse json for config %s failed: %s", cfg, cc_last_error());
			free(old_cfgstr);
			free(new_cfgstr);
			record_error(_F("{1} format error", "JSON"), _T(cfg), NULL);
			return CC_INVALID_PARAM;
		}

		/* 使用新的json设置到配置中，调用cc_cache_set_cfg_json接口后，newjs将被释放掉 */
		// 函数内部记录错误值
		idl_object_st *newobj = convert_upgrade_json_to_config(cfg, newjs);
		if (!newobj) {
			free(old_cfgstr);
			free(new_cfgstr);
			idl_object_release(newobj);
			return -1;
		}

		if ((res = cc_cdb_config_set(cdb, cfg, newobj))) {
			idl_object_release(newobj);
			record_error(cc_last_error(), _T(cfg), NULL);
			return res;
		}

		debug("Upgrade config %s success.", cfg);

		/* 删掉老文件 */
		str_snprintf(cfgpath, PATH_MAX, "%s/%s/%s/%s.base", cc_cache_db_path(), tenant, db, cfg);
		unlink(cfgpath);
		str_snprintf(cfgpath, PATH_MAX, "%s/%s/%s/%s.actions", cc_cache_db_path(), tenant, db, cfg);
		unlink(cfgpath);
		str_snprintf(cfgpath, PATH_MAX, "%s/%s/%s/%s.digest", cc_cache_db_path(), tenant, db, cfg);
		unlink(cfgpath);

		free(old_cfgstr);
		free(new_cfgstr);
		idl_object_release(newobj);
	}

	if ((res = cc_cdb_commit(cdb))) {
		log_debug("cfg-center", "upgrader", "Set cdb for tenant(%s) db(%s) failed: %s", tenant, db, cc_last_error());
		record_error(_T("Write config failed"), NULL, NULL);
		return res;
	}

	/* CDB文件保存成功后，需要清除AOW */
	cc_aow_clear(tenant, db);

	/* 执行配置升级后置处理器 */
	res = cc_upgradposter_call(tenant, db);
	if (res) {
		debug("Execute upgradeposter failed: %s", cc_last_error());
		record_error(_T("Validation post check failed"), NULL, NULL);
		return res;
	}

	/* 删除原持久化目录 */
	str_snprintf(cfgpath, PATH_MAX, "%s/%s/persistent", cc_cache_db_path(), tenant);
	dirutils_rmdir(cfgpath);

	return 0;
}

#include <unistd.h>
#include "rt_shm.h"
#include "rt_shm_monitor_deal.h"
#include "afdir/afdir_safe.h"
#include "ctrlpanel/imdtsk_clt.h"
/**
 * @brief 销毁共享内存，由sase_monitor调用
 * @return: 摧毁成功返回1，不摧毁返回0
 */
int sase_rt_monitor_destory()
{
    if (!sase_rt_shm_is_ready()) {
//        sase_rt_shm_destory();
        return 0;
    }

    sase_rt_agent_info_shm_t *sase_rt_agent_info_shm = get_sase_rt_agent_info_shm();
    sase_rt_agent_info_shm->init_suc = 0;

    sleep(5);

    sase_rt_shm_destory();

    return 1;
}

idl_object_st *idl_object_from_net_overload_protection(idl_mgr_st *mgr, const net_overload_protection *obj)
{
	schema_st *sch = schema_mgr_get(idl_file_get_schema_mgr(IDL_MF(mgr)), "net.overload.protection");
	if (!sch)
		return NULL;
	
	idl_object_st *ret = idl_object_new(sch);
	idl_object_st *fobj = NULL;

	if (NET_OVERLOAD_PROTECTION_HAS_FIELD(obj, ENABLE)) {
		fobj = idl_object_from_net_overload_protection_enable(mgr, obj->enable);
		if (!fobj) 
			goto failed_ret;
		
		if (idl_object_set_field(ret, "enable", fobj))
			goto failed_fobj;
		
		idl_object_release(fobj);
	}

	return ret;

failed_fobj:
	idl_object_release(fobj);
failed_ret:
	idl_object_release(ret);
	return NULL;
}

/**
 * 策略路由搜索树更新添加
 * @param [in] msg_pbr  消息结构体
 * @param [in] pbr  策略路由结构体
 * @return 0 成功
 * @return !0 失败
 */
sf_int32_t dp_pbr_search_add_for_update(sf_pbr_t* p_pbr_info, 
                                                      dp_pbr_entry_t *pbr)
{
    assert(p_pbr_info && pbr);

    if (sf_pbr_update_is_src_zone(pbr)){
        if (dp_pbr_search_add_src_zone(p_pbr_info->pbr_search, pbr) == FALSE) {
            log_debug("dp pbr", "", "[error]:pbr %s search add src zone fail!\n", 
                     pbr->pbr_name);
            goto SRCZONE_ADD_FAIL;
        }
    }

    if (sf_pbr_update_is_srcip(pbr)) {
        if (dp_pbr_search_add_src_ip(p_pbr_info->pbr_search, pbr) == FALSE) {
            log_debug("dp pbr", "", "[error]:pbr %s search add src ip fail!\n", 
                     pbr->pbr_name);
            goto SRCIP_ADD_FAIL;
        }
    }

    if (sf_pbr_update_is_dstip(pbr)) {
        if (dp_pbr_search_add_dst(p_pbr_info, pbr) == FALSE){
            log_debug("dp pbr", "", "[error]:pbr %s search add dst fail!\n", pbr->pbr_name);
            goto DST_ADD_FAIL;
        }
    }

    if (sf_pbr_update_is_service(pbr)) {
        if (dp_pbr_search_add_service(p_pbr_info->pbr_search, pbr) == FALSE) {
            log_debug("dp pbr", "", "[error]:pbr %s search add service fail!\n", pbr->pbr_name);
            goto SERVICE_ADD_FAIL;
        }
    }

    if (sf_pbr_update_is_app(pbr)){
        if (dp_pbr_search_add_app(p_pbr_info->pbr_search, pbr) == FALSE) {
            log_debug("dp pbr", "", "[error]:pbr %s search add app fail\n", pbr->pbr_name);
            goto APP_ADD_FAIL;
        }
    }

    if (pbr->rule_type == PBR_IPV4) {
        if (sf_pbr_update_is_user(pbr)){
            if (dp_pbr_search_add_user(p_pbr_info->pbr_search, pbr) == FALSE) {
                log_debug("dp pbr", "", "[error]:pbr %s search add user fail\n", pbr->pbr_name);
                goto USER_ADD_FAIL;
            }
        }

        if (sf_pbr_update_is_usergroup(pbr)){
            if (dp_pbr_search_add_usergroup(p_pbr_info->pbr_search, pbr) == FALSE) {
                log_debug("dp pbr", "", "[error]:pbr %s search add usergroup fail\n", pbr->pbr_name);
                goto USERGROUP_ADD_FAIL;
            }
        }
    }

    return 0;

USERGROUP_ADD_FAIL:
    if (sf_pbr_update_is_user(pbr)) {
        if (dp_pbr_search_delete_user(p_pbr_info->pbr_search, pbr) == FALSE) {
            log_debug("dp pbr", "", "[error]:pbr %s search tree delete user fail\n", pbr->pbr_name);
        }
    }

USER_ADD_FAIL:
    if (sf_pbr_update_is_app(pbr)) {
        if (dp_pbr_search_delete_app(p_pbr_info->pbr_search, pbr) == FALSE) {
            log_debug("dp pbr", "", "[error]:pbr %s search tree delete app fail\n", pbr->pbr_name);
        }
    }

APP_ADD_FAIL:
    if (sf_pbr_update_is_service(pbr)) {
        if (dp_pbr_search_delete_service(p_pbr_info->pbr_search, pbr) == FALSE) {
            log_debug("dp pbr", "", "[error]:pbr id %u search tree delete dst service fail\n", 
                         pbr->pbr_id);
        }
    }
SERVICE_ADD_FAIL:
    if (sf_pbr_update_is_dstip(pbr)) {
        if (dp_pbr_search_delete_dst(p_pbr_info, pbr) == FALSE) {
            log_debug("dp pbr", "", "[error]:pbr id %u search tree delete dst fail\n", 
                         pbr->pbr_id);
        }
    }
DST_ADD_FAIL:
    if (sf_pbr_update_is_srcip(pbr)) {
        if (dp_pbr_search_delete_src_ip(p_pbr_info->pbr_search, pbr) == FALSE) {
            log_debug("dp pbr", "", "[error]:pbr id %u search tree delete src ip fail\n", 
                          pbr->pbr_id);
        }
    }
SRCIP_ADD_FAIL:
   if (sf_pbr_update_is_src_zone(pbr)){
        if (dp_pbr_search_delete_src_zone(p_pbr_info->pbr_search, pbr) == FALSE) {
            sf_debug(PBR, "pbr id %u search delete src zone fail\n", 
                     pbr->pbr_id);
            goto SRCZONE_ADD_FAIL;
        }
    }
SRCZONE_ADD_FAIL:
    return -1;
}

bool cellularif_cmd_move(cellularif_cmd_t* lvp, cellularif_cmd_t* rvp)
{
    if (lvp == NULL || rvp == NULL) {
        return false;
    }

    if (!cellularif_cmd_clear(lvp))
        return false;

    lvp->cmd = rvp->cmd;
    lvp->arg = rvp->arg;

    rvp->cmd = CELLULARIF_CMD_NONE;
    rvp->arg.update_status = 0;

    return true;
}

void net_pbr4Status_srcIpGroups_destroy(net_pbr4Status_srcIpGroups *obj)
{
	if (!obj)
		return;

	uint32_t i;
	for (i = 0; i < obj->count; ++i) {
		net_pbr4Status_srcIpGroups_items_destroy(obj->items[i]);
	}
	
	free(obj->items);
	free(obj);
}

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

void net_pbr4_srcIpGroups_destroy(net_pbr4_srcIpGroups *obj)
{
	if (!obj)
		return;

	uint32_t i;
	for (i = 0; i < obj->count; ++i) {
		net_pbr4_srcIpGroups_items_destroy(obj->items[i]);
	}
	
	free(obj->items);
	free(obj);
}

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

void net_pbr4_dstIsp_destroy(net_pbr4_dstIsp *obj)
{
	if (!obj)
		return;

	uint32_t i;
	for (i = 0; i < obj->count; ++i) {
		net_pbr4_dstIsp_items_destroy(obj->items[i]);
	}
	
	free(obj->items);
	free(obj);
}

net_pbr4Status_dstPort *net_pbr4Status_dstPort_from_idl_object(idl_object_st *obj)
{
	if (!obj || strcmp(schema_get_gid(idl_object_get_schema(obj)), "net.pbr4Status.dstPort") != 0)
		return NULL;
	
	int nr_elem = idl_object_get_length(obj);
	if (nr_elem < 0)
		return NULL;
	
	net_pbr4Status_dstPort *ret = zero_alloc(sizeof(net_pbr4Status_dstPort));
	ret->_digest = idl_object_get_digest(obj);
	ret->items = alloc_die(nr_elem * sizeof(portRange *));
	
	int i;
	for (i = 0; i < nr_elem; ++i) {
		idl_object_st *item = NULL;
		idl_object_get_item(obj, i, &item);
		ret->items[i] = portRange_from_idl_object(item);
	}
	
	ret->count = (uint32_t)nr_elem;
	ret->_size = (uint32_t)nr_elem;
	return ret;
}

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