import type { HttpClient } from '../../../../shared/api-client/http';
import type { StatusDTO } from '../../../../shared/api-client/generated/onboarding';
import type { UpdateDTO } from '../../../../shared/api-client/generated/onboarding-requests';
import type { OnboardingGateway, OnboardingStatus } from '../../usecase/onboardingUseCases';

/** OnboardingGateway backed by the REST API. */
export class OnboardingApi implements OnboardingGateway {
  constructor(private readonly http: HttpClient) {}

  async status(): Promise<OnboardingStatus> {
    const { completed, step } = await this.http.get<StatusDTO>('/api/onboarding');
    return { completed, step };
  }

  async setCompleted(completed: boolean): Promise<void> {
    const body: UpdateDTO = { completed };
    await this.http.put<StatusDTO>('/api/onboarding', body);
  }

  async setStep(step: string): Promise<void> {
    const body: UpdateDTO = { completed: false, step };
    await this.http.put<StatusDTO>('/api/onboarding', body);
  }
}
