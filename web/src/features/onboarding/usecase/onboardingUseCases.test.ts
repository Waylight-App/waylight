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
  it('starts the wizard for a fresh account and continues where it stopped', async () => {
    expect(await setup({ completed: false, step: '' }).uc.state()).toEqual({ stage: 'wizard', step: '' });
    expect(await setup({ completed: false, step: 'widgets' }).uc.state()).toEqual({
      stage: 'wizard',
      step: 'widgets',
    });
  });

  it('shows the tour after the wizard and nothing after both', async () => {
    expect((await setup({ completed: false, step: 'done' }).uc.state()).stage).toBe('tour');
    expect((await setup({ completed: true, step: '' }).uc.state()).stage).toBe('finished');
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
