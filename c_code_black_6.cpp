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