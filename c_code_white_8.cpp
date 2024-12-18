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