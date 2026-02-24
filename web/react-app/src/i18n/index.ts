import { en, TranslationKeys } from './en';
import { zh } from './zh';

export type Language = 'en' | 'zh';

export const translations: Record<Language, TranslationKeys> = {
  en,
  zh,
};

export type { TranslationKeys };
export { en, zh };
