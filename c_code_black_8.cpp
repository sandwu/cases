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