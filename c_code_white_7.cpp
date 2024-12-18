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