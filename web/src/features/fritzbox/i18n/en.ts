/** English UI texts of the fritzbox feature added with the calls widget (#168, #193). */
export const en = {
  'widget.fritzbox_calls.title': 'FRITZ!Box calls',
  'widget.fritzbox_calls.description': 'Recent calls of the FRITZ!Box: all, missed, incoming or outgoing.',
  'fritzbox.show_uptime': 'Show the uptime',
  'fritzbox.show_ip': 'Show the external IP address',
  'fritzbox.show_rates': 'Show the current download and upload',
  'fritzbox.calls_filter': 'Calls',
  'fritzbox.filter.all': 'All calls',
  'fritzbox.filter.missed': 'Missed',
  'fritzbox.filter.incoming': 'Incoming',
  'fritzbox.filter.outgoing': 'Outgoing',
  'fritzbox.calls_count': 'Number of calls (5–50)',
  'fritzbox.show_numbers': 'Show numbers of callers without a name',
  'fritzbox.save': 'Save',
  'fritzbox.no_calls': 'No calls.',
  'fritzbox.call.incoming': 'incoming',
  'fritzbox.call.missed': 'missed',
  'fritzbox.call.outgoing': 'outgoing',
  'fritzbox.call.rejected': 'rejected',
  'fritzbox.call.active': 'active',
  'fritzbox.setup_username': 'FRITZ!Box user',
  'fritzbox.setup_password': 'Password',
  'fritzbox.setup_step_network':
    'Open the FRITZ!Box, go to Home Network → Network → Network Settings and tick Allow access for applications and Transmit status information over UPnP.',
  'fritzbox.setup_step_user':
    'For the line speed and the call list: under System → FRITZ!Box Users add a user Waylight with the right FRITZ!Box Settings.',
  'fritzbox.setup_step_user_calls':
    'Under System → FRITZ!Box Users add a user Waylight with the rights FRITZ!Box Settings and Voice messages, faxes, FRITZ!App Fon and call list.',
  'fritzbox.setup_step_enter_optional': 'Enter that user and its password below, or leave both empty.',
  'fritzbox.setup_step_enter': 'Enter that user and its password below.',
  'fritzbox.setup_online': 'Online, external address {ip}',
  'fritzbox.setup_offline': 'The FRITZ!Box answers; the internet connection is down.',
  'fritzbox.setup_calls': '{n} recent calls',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
