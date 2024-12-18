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