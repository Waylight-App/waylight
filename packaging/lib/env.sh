# Shared by the waylight helper scripts (sourced, not executed).
#
# waylight_value VAR prints the value the service sees for VAR, following
# the order of EnvironmentFile= in the unit: waylight.env (admin) wins over
# appliance.env (Raspberry Pi image defaults), which wins over secret.env.

WAYLIGHT_CONF_DIR=${WAYLIGHT_CONF_DIR:-/etc/waylight}

waylight_value() {
	for f in waylight.env appliance.env secret.env; do
		v=$(sed -n "s/^[[:space:]]*$1=\([^[:space:]]*\).*/\1/p" "$WAYLIGHT_CONF_DIR/$f" 2>/dev/null | tail -n 1)
		if [ -n "$v" ]; then
			echo "$v"
			return
		fi
	done
}

# waylight_url_suffix prints ":<port>" unless the service listens on 80.
waylight_url_suffix() {
	listen=$(waylight_value LISTEN_ADDR)
	port=${listen##*:}
	case "$port" in
	"" | 80) ;;
	*) echo ":$port" ;;
	esac
}

# waylight_setup_status prints pending, done or unknown (service not up).
waylight_setup_status() {
	listen=$(waylight_value LISTEN_ADDR)
	env LISTEN_ADDR="${listen:-:8090}" waylight setup-status 2>/dev/null || echo unknown
}
