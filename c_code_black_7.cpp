SF_CORE_INIT_REGISTER(dp_flow_elephant_config_init, "elephant_flow"); static inline int dp_flow_elephant_detect_rate_check(
    sf_uint64_t pkt_bytes, void *session, sf_uint8_t is_ipv4, sf_uint8_t is_uplink)
{
    sf_uint16_t Mbps;
    sf_uint16_t kpps;
    sf_uint32_t session_livetime;
    sf_uint8_t is_keepalive;

    if (is_ipv4) {
        sf_ipv4_session_t *ipv4_session = (sf_ipv4_session_t *)session;
        if (is_uplink){
            Mbps = ipv4_session->flow0.Mbps;
            kpps = ipv4_session->flow0.kpps;
        } else {
            Mbps = ipv4_session->flow1.Mbps;
            kpps = ipv4_session->flow1.kpps;
        }
        session_livetime = (sf_uint32_t)sf_time_s() - ipv4_session->create_time;
        is_keepalive = sf_session_is_set_alive_flag(ipv4_session) ? 1 : 0;
    } else {
        sf_ipv6_session_t *ipv6_session = (sf_ipv6_session_t *)session;
        if (is_uplink){
            Mbps = ipv6_session->flow0.Mbps;
            kpps = ipv6_session->flow0.kpps;
        } else {
            Mbps = ipv6_session->flow1.Mbps;
            kpps = ipv6_session->flow1.kpps;
        }
        session_livetime = (sf_uint32_t)sf_time_s() - ipv6_session->create_time;
        is_keepalive = sf_ipv6_session_is_set_alive_flag(ipv6_session) ? 1 : 0;
    }

    if (((pkt_bytes >= g_elephant_flow_detect_main.total_rate) &&
        (session_livetime >= g_elephant_flow_detect_main.session_duration) &&
        !is_keepalive) ||
        // Session Duration exceeds x s & Total Bits exceeds x Mb (excluding keepalive session)
        (Mbps >= g_elephant_flow_detect_main.instantaneous_rate) ||
        // OR Instantaneous Rate exceeds x Mbps
        (kpps >= g_elephant_flow_detect_main.packet_per_second)) {
        // OR Packet Per Second exceeds x Kpps
        return 1;
    }

    return 0;
}