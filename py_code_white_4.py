def write_subnet(account_id, vpc_vid, cidr_pos):
    row = subnet_row(account_id, cidr_pos)
    subnet_fd.writerow((row[0], row[1]))
    subnet_vid = row[0]
    subnet_id = row[1]

    util.write_both_access(subnet_vpc_fd, vpc_vid, subnet_vid)

    db.add_rule_asset({
        **db.template_rule_asset_not_admin_id,
        "_id": row[0],
        "assetId": row[0],
        "assetIdShow": row[2],
        "assetName": row[1],
        "assetType": 20,
        "cloudSource": 1,
        "cloudAccountId": account_id,
        "data": {
            "region": {
                "regionId": "region",
                "regionName": "亚太-新加坡",
            },
            "description": "subnet",
            "cidr": "192.168.1.0/24",
            "gatewayIp": "192.168.1.254",
            "vpc": {
                "relationVpcId": vpc_vid,
                "vpcId": vpc_vid,
                "vpcName": vpc_vid,
            },
            "status": "ACTIVE",
            "routeTableId": "",
            "routeTable": {
                "status": "",
            },
        }
    })
    return subnet_vid, subnet_id