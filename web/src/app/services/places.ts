import { PlacesApi } from '../../features/places/adapter/api/placesApi';
import { PlacesUseCases } from '../../features/places/usecase/placesUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    places: PlacesUseCases;
  }
}

// Direct imports keep UI exports of the feature index out of the initial bundle (#230).
export const service = defineService('places', (http) => new PlacesUseCases(new PlacesApi(http)));
