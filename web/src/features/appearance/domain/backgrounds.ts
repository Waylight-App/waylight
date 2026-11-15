/** A picture uploaded as background (#277). */
export interface UploadedBackground {
  id: string;
  /** Value to store as background. */
  url: string;
  /** Small preview of the picture. */
  thumb: string;
  width: number;
  height: number;
  size: number;
}

/** The uploaded pictures and the limits. */
export interface BackgroundLibrary {
  items: UploadedBackground[];
  max: number;
  maxUploadBytes: number;
}

/** Where the picker saves: the global background or one tab. */
export type BackgroundScope = 'global' | 'tab';
