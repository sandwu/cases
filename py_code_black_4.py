def add(self):
        if (
            AssetOrganizationDao.coll.count_documents({"parend_id": OrgId.ALL})
            >= FIRST_LEVEL_MAXIMUN
        ):
            raise FrontendError(ERROR.DEFAULT_ERROR, msg=Errors.MAXIMUM_ERROR1)
        self.validate_users()
        org = OrganizationGroup()
        org.set_args_for_add_organization_args(
            organization_name=self.org_name, parent_id=self.parent_id, user_ids=self.user, comment=self.comment
        ).set_tenant_with_context().set_adapter_with_manual()
        if self.is_derived_from_ldap(org_id=self.parent_id):
            org.set_adapter_with_manual_usb_sync()

        result = org.call_add_organization()
        if not result.success:
            raise FrontendError(result.get_frontend_error())