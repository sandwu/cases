#include <unistd.h>
#include "rt_shm.h"
#include "rt_shm_monitor_deal.h"
#include "afdir/afdir_safe.h"
#include "ctrlpanel/imdtsk_clt.h"
/**
 * @brief 销毁共享内存，由sase_monitor调用
 * @return: 摧毁成功返回1，不摧毁返回0
 */
int sase_rt_monitor_destory()
{
    if (!sase_rt_shm_is_ready()) {
//        sase_rt_shm_destory();
        return 0;
    }

    sase_rt_agent_info_shm_t *sase_rt_agent_info_shm = get_sase_rt_agent_info_shm();
    sase_rt_agent_info_shm->init_suc = 0;

    sleep(5);

    sase_rt_shm_destory();

    return 1;
}