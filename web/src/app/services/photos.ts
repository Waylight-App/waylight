import { PhotosApi } from '../../features/photos/adapter/api/photosApi';
import { PhotosUseCases } from '../../features/photos/usecase/photosUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    photos: PhotosUseCases;
  }
}

// Direct imports keep UI exports of the feature index out of the initial bundle (#230).
export const service = defineService('photos', (http) => new PhotosUseCases(new PhotosApi(http)));
