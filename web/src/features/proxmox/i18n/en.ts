/** English UI texts of the proxmox feature (#193). */
export const en = {
  'widget.proxmox.title': 'Proxmox VE',
  'widget.proxmox.description': 'Node load and the status of VMs and LXC containers.',
  'settings.section.proxmox': 'Proxmox',
  'proxmox.connect_under_settings': 'Your servers and machines from Proxmox.',
  'proxmox.uptime': 'up {value}',
  'proxmox.offline': 'offline',
  'proxmox.cpu': 'CPU ({cpus} cores)',
  'proxmox.memory': 'Memory',
  'proxmox.disk': 'Root disk',
  'proxmox.no_guests': 'No VMs or containers.',
  'proxmox.status.stopped': 'stopped',
  'proxmox.status.paused': 'paused',
  'proxmox.status.suspended': 'suspended',
  'proxmox.status.prelaunch': 'starting',
  'proxmox.status.unknown': 'unknown',
  'proxmox.start': 'Start {name}',
  'proxmox.shutdown': 'Shut down {name}',
  'proxmox.confirm_start': 'Start {name}?',
  'proxmox.confirm_shutdown': 'Shut down {name}? Running services in it will stop.',
  'proxmox.running_only': 'Only show running guests',
  'proxmox.url': 'Proxmox VE address',
  'proxmox.token_id': 'API token ID',
  'proxmox.secret': 'Token secret {value}',
  'proxmox.fingerprint': 'Certificate fingerprint (SHA-256, optional)',
  'proxmox.fingerprint_hint':
    'Proxmox uses a self-signed certificate. Paste its SHA-256 fingerprint (node → System → Certificates) to trust exactly this certificate, or leave it empty if the certificate is signed by a known CA.',
  'proxmox.token_hint':
    'Create a read-only API token (role PVEAuditor) on a Proxmox node, then enter the token ID and secret below.',
  'proxmox.actions_hint':
    'Start and shutdown stay off unless PROXMOX_ACTIONS_GUESTS lists the guest IDs and the token may power them (VM.PowerMgmt).',
  'proxmox.setup_token_id': 'Token ID',
  'proxmox.setup_secret': 'Secret',
  'proxmox.setup_step_open': 'Open Proxmox VE and go to Datacenter → Permissions → API Tokens.',
  'proxmox.setup_step_add':
    'Press Add, choose a user, name the token and untick Privilege Separation, or give the token its own role.',
  'proxmox.setup_step_role':
    'The token needs read access (PVEAuditor) on /; start and shutdown also need VM.PowerMgmt.',
  'proxmox.setup_step_copy':
    'Copy the token ID and the secret and paste them below; the secret is shown only once.',
  'proxmox.setup_found': '{nodes} nodes, {guests} guests',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
