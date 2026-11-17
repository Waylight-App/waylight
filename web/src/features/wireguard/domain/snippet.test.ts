import { describe, expect, it } from 'vitest';
import { exportCronSnippet } from './snippet';

describe('exportCronSnippet', () => {
  it('uses the default container without extra variables', () => {
    expect(exportCronSnippet({ stackDir: '/opt/stacks/waylight/', container: 'wireguard' })).toBe(
      "sudo tee /etc/cron.d/waylight-wireguard <<'CRON'\n* * * * * root /opt/stacks/waylight/scripts/wireguard-export.sh\nCRON",
    );
  });

  it('sets WG_CONTAINER for other containers and for wg on the host', () => {
    expect(exportCronSnippet({ stackDir: '/srv/th', container: 'wg-easy' })).toContain(
      'root WG_CONTAINER=wg-easy /srv/th/scripts',
    );
    expect(exportCronSnippet({ stackDir: '/srv/th', container: '' })).toContain('root WG_CONTAINER= /srv/th');
  });

  it('rejects relative paths and shell metacharacters', () => {
    expect(exportCronSnippet({ stackDir: 'stacks', container: 'wireguard' })).toBeNull();
    expect(exportCronSnippet({ stackDir: '/opt/$(id)', container: 'wireguard' })).toBeNull();
    expect(exportCronSnippet({ stackDir: '/opt', container: 'a;b' })).toBeNull();
  });
});
