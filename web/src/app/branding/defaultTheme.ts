import type { HttpClient } from '../../shared/api-client/http';
import type { DefaultThemeDTO } from '../../shared/api-client/generated/appearance';

/**
 * The instance's default theme (DEFAULT_THEME, #507) for the sign-in page.
 * After sign-in the account's appearance comes from GET /api/appearance,
 * which already includes the default theme for accounts without their own,
 * so the style is only needed while nobody is signed in.
 */
export async function loadDefaultTheme(http: HttpClient): Promise<DefaultThemeDTO> {
  try {
    return await http.get<DefaultThemeDTO>('/api/appearance/default-theme');
  } catch {
    // An older server or a network error: the built-in look stays.
    return { id: '', css: '', scheme: '' };
  }
}

/**
 * Puts the theme's CSS into the page and returns a function that takes it
 * out again. textContent, not innerHTML, so the style sheet cannot carry
 * markup; the server builds it from validated tokens only.
 */
export function showDefaultTheme(theme: DefaultThemeDTO, doc: Document = document): () => void {
  if (!theme.css) return () => undefined;
  const style = doc.createElement('style');
  style.dataset['source'] = 'default-theme';
  style.textContent = theme.css;
  doc.head.append(style);
  return () => style.remove();
}
