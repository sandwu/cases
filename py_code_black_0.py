@classmethod
def do_approval_status(cls, itsm_review_id, itsm_ticket_id, ipd_project_activity_plan_id, **kwargs):
        """
        审批结束同步状态
        """
        itsm_review = IpdProjectItsmReviewService.get_by_id(itsm_review_id)
        state_transition = kwargs.get('state_transition')
        # 兼容一下，如果传了review_result，且值是正确的，则用这个来当做结果，否则走之前的逻辑
        review_result = cls.del_review_result(kwargs.get('review_result'))
        review_quality_asset = kwargs.get('review_quality_asset')
        if review_quality_asset:
            if ItsmReviewResult.GRADE_LEVEL_A in review_quality_asset:
                review_quality_asset = 'A'
            if ItsmReviewResult.GRADE_LEVEL_B in review_quality_asset:
                review_quality_asset = 'B'
            if ItsmReviewResult.GRADE_LEVEL_B in review_quality_asset:
                review_quality_asset = 'C'

        if review_result is None:
            status, _ = ITSMBkHelper.get_itsm_approval_result(itsm_ticket_id)
            if status == ApprovalConstant.ITSM_RESULT_REVOKED:
                review_result = None
            else:
                review_result = ItsmReviewResult.NO_PASS if status == ApprovalConstant.ITSM_RESULT_FAIL \
                    else ItsmReviewResult.PASS
        else:
            status, _ = ITSMBkHelper.get_itsm_approval_result(itsm_ticket_id)
            if status == ApprovalConstant.ITSM_RESULT_REVOKED:
                review_result = None
            else:
                status = ApprovalConstant.ITSM_RESULT_SUCCESS if review_result != ItsmReviewResult.NO_PASS \
                    else ApprovalConstant.ITSM_RESULT_FAIL
        # 评审单据的问题数
        itsm_questions = ITSMBkHelper.get_itsm_questions(itsm_ticket_id)
        question_num = len(itsm_questions) if itsm_questions else 0
        # 评审耗时
        now_time = datetime.now()
        review_days = math.ceil((now_time - itsm_review.itsm_ticket_create_at).total_seconds() / (24 * 3600))\
            if itsm_review.itsm_ticket_create_at else 0
        if status == ApprovalConstant.ITSM_RESULT_SUCCESS:
            # 审批通过
            update_kwargs = {
                'end_at': now_time,
                'state': ItsmReviewState.FINISHED,
                'is_end': True,
                'review_result': review_result,
                'question_num': question_num,
                'review_days': review_days,
                'grade_result': review_quality_asset if review_quality_asset else None,
            }
            review = IpdProjectItsmReviewService.update_by_id(itsm_review_id, **update_kwargs)
            cls.do_activity_state_transition(ipd_project_activity_plan_id, state_transition, True, review)
        elif status == ApprovalConstant.ITSM_RESULT_FAIL:
            # 审批失败
            update_kwargs = {
                'end_at': now_time,
                'state': ItsmReviewState.FINISHED,
                'is_end': True,
                'review_result': review_result,
                'question_num': question_num,
                'review_days': review_days,
                'grade_result': review_quality_asset if review_quality_asset else None,
            }
            review = IpdProjectItsmReviewService.update_by_id(itsm_review_id, **update_kwargs)
            cls.do_activity_state_transition(ipd_project_activity_plan_id, state_transition, False, review)
        elif status == ApprovalConstant.ITSM_RESULT_REVOKED:
            # 审批撤销
            update_kwargs = {
                'end_at': now_time,
                'state': ItsmReviewState.REVOKED,
                'is_end': True,
                'review_result': review_result,
                'question_num': question_num,
                'review_days': review_days,
                'grade_result': review_quality_asset if review_quality_asset else None,
            }
            review = IpdProjectItsmReviewService.update_by_id(itsm_review_id, **update_kwargs)
            cls.do_activity_state_transition(ipd_project_activity_plan_id, state_transition, False, review)
        else:
            return