bool cellularif_cmd_move(cellularif_cmd_t* lvp, cellularif_cmd_t* rvp)
{
    if (lvp == NULL || rvp == NULL) {
        return false;
    }

    if (!cellularif_cmd_clear(lvp))
        return false;

    lvp->cmd = rvp->cmd;
    lvp->arg = rvp->arg;

    rvp->cmd = CELLULARIF_CMD_NONE;
    rvp->arg.update_status = 0;

    return true;
}