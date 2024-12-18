def __night_accept_rate_get(self, custom, day_ago_format):
        # 查找推荐的告警
        rec_result = self._rec_collection.find({'handle_time': {'$gte': day_ago_format},
                                                "company_id": custom})
        # 查找晚上推荐被驳回的数量，定义晚上6点和第二天9点的时间
        start_time = time(18, 0, 0)
        end_time = time(9, 0, 0)
        total_num_night = 0
        accept_num_night = 0
        for result in rec_result:
            alert_id = result.get(AlertConst.ID)
            alert = self._history_collection.find_one({AlertConst.ID: ObjectId(alert_id)})
            if not alert:
                continue

            # 判断推荐告警是否有效
            event_status = alert.get(AlertConst.EVENT_STATUS, '')
            # 使用推荐的handle_time，原始告警的handle_time相差8小时
            handle_time = result.get('handle_time')
            # 判断黑白事件
            if event_status not in AlertConst.BLACK_EVENT and \
                    event_status not in AlertConst.WHITE_EVENT:  # 推荐结果未出不考虑
                continue

            handle_strptime = datetime.strptime(handle_time, RecConst.RecFormatDateTime)
            # 判断时间是否在晚上6点到第二天9点之间
            handle_time = handle_strptime.time()
            if start_time <= handle_time or handle_time <= end_time:
                total_num_night += 1
                if event_status in AlertConst.BLACK_EVENT:
                    accept_num_night += 1

        if total_num_night == 0:
            return 1

        return accept_num_night / total_num_night