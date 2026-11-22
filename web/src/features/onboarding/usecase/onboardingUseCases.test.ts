import { describe, expect, it, vi } from 'vitest';
import { OnboardingUseCases, type OnboardingGateway } from './onboardingUseCases';

function setup(status: { completed: boolean; step: string }) {
  const gateway: OnboardingGateway = {
    status: vi.fn().mockResolvedValue(status),
    setCompleted: vi.fn().mockResolvedValue(undefined),
    setStep: vi.fn().mockResolvedValue(undefined),
  };
  return { gateway, uc: new OnboardingUseCases(gateway) };
}

describe('OnboardingUseCases', () => {
  it('reads the stored state', async () => {
    expect(await setup({ completed: false, step: 'widgets' }).uc.status()).toEqual({
      completed: false,
      step: 'widgets',
    });
  });

  it('restarts the tour alone, past the wizard', async () => {
    const { uc, gateway } = setup({ completed: true, step: '' });
    await uc.restart();
    expect(gateway.setStep).toHaveBeenCalledWith('done');
  });

  it('completes the tour', async () => {
    const { uc, gateway } = setup({ completed: false, step: 'done' });
    await uc.finish();
    expect(gateway.setCompleted).toHaveBeenCalledWith(true);
  });
});
