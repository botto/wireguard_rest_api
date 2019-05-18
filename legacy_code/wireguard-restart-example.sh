ip netns exec your_namespace wg-quick down "$WG_CONFIG_NAME"
ip netns exec your_namespace wg-quick up "$WG_CONFIG_NAME"
