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