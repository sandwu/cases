void dp_overload_protection_enable_show(struct idl_param_st *p, struct idl_buf_st *output, void *userdat)
{
    struct dp_overload_context_t *ctx = dp_overload_ctx();
    assert(ol);
    if (output != NULL) {
        idl_buf_printf(output, DP_METRICS_STR_MAX, "%s", ctx->cfg.enable ? "enable" : "disable");
    }
}