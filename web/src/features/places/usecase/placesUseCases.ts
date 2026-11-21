import type { SavedPlace, SavedPlaces } from '../../../shared/places/places';
import type { Place, PlaceCandidate } from '../domain/places';

/** Output port: the places backend. */
export interface PlacesGateway {
  list(): Promise<Place[]>;
  add(p: PlaceCandidate): Promise<Place>;
  update(id: string, p: PlaceCandidate): Promise<Place>;
  remove(id: string): Promise<void>;
}

/**
 * Application logic of saved places (#591). It is also the SavedPlaces
 * port widget settings use to choose and save a place.
 */
export class PlacesUseCases implements SavedPlaces {
  constructor(private readonly gateway: PlacesGateway) {}

  list(): Promise<Place[]> {
    return this.gateway.list();
  }

  add(p: Omit<SavedPlace, 'id'>): Promise<Place> {
    return this.gateway.add(trimmed(p));
  }

  /** Renames a place, keeping where it is. */
  rename(place: Place, name: string): Promise<Place> {
    return this.gateway.update(place.id, trimmed({ ...place, name }));
  }

  remove(id: string): Promise<void> {
    return this.gateway.remove(id);
  }
}

function trimmed(p: PlaceCandidate): PlaceCandidate {
  return { name: p.name.trim(), label: p.label.trim(), lat: p.lat, lon: p.lon };
}
