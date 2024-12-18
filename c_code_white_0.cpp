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