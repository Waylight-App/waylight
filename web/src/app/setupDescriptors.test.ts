import { describe, expect, it } from 'vitest';
import { arrSetup } from '../features/arr/ui/arrSetup';
import { adguardSetup } from '../features/adguard/ui/adguardSetup';
import { calendarSetup } from '../features/calendar/ui/calendarSetup';
import { fritzboxSetup } from '../features/fritzbox/ui/fritzboxSetup';
import { fuelSetup } from '../features/fuel/ui/fuelSetup';
import { grafanaSetup } from '../features/grafana/ui/grafanaSetup';
import { immichSetup } from '../features/immich/ui/immichSetup';
import { mediaSetup } from '../features/media/ui/mediaSetup';
import { nextcloudSetup } from '../features/nextcloud/ui/nextcloudSetup';
import { proxmoxSetup } from '../features/proxmox/ui/proxmoxSetup';
import { shoppingSetup } from '../features/shopping/ui/shoppingSetup';
import { smarthomeSetup } from '../features/smarthome/ui/smarthomeSetup';
import { speedtestSetup } from '../features/speedtest/ui/speedtestSetup';
import { fieldHint, pasteTarget, type SetupDescriptor } from '../shared/setup/setup';
import { t } from '../shared/i18n';

// The use cases are only called by check(), which these tests never run.
const stub = {} as never;
const descriptors: Record<string, SetupDescriptor> = {
  arr: arrSetup(stub),
  'adguard home': adguardSetup(stub, 'adguard'),
  pihole: adguardSetup(stub, 'pihole'),
  calendar: calendarSetup(stub),
  fritzbox: fritzboxSetup(stub, false),
  'fritzbox calls': fritzboxSetup(stub, true),
  fuel: fuelSetup(stub, () => Promise.resolve([])),
  grafana: grafanaSetup(stub),
  immich: immichSetup(stub),
  media: mediaSetup(stub),
  nextcloud: nextcloudSetup(stub),
  proxmox: proxmoxSetup(stub),
  shopping: shoppingSetup(stub),
  smarthome: smarthomeSetup(stub),
  speedtest: speedtestSetup(stub),
};

describe('the guided setup of every integration (#738)', () => {
  for (const [name, d] of Object.entries(descriptors)) {
    it(`${name} tells where to find every value`, () => {
      const secret = pasteTarget(d);
      expect(secret, 'a key field the phone can fill').toBeDefined();
      // The key's field shows the service's steps; every other field has a note.
      expect(d.steps.length).toBeGreaterThan(0);
      for (const f of d.fields.filter((x) => x !== secret)) {
        expect(fieldHint(f, d.service), f.key).not.toBe('');
      }
    });
  }

  it('writes the address note with the product name', () => {
    const f = descriptors['immich']?.fields[0];
    expect(f && fieldHint(f, 'Immich')).toBe(t('setup.address_hint', { service: 'Immich' }));
  });
});
