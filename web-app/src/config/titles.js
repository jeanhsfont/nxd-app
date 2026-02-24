/**
 * Única fonte de verdade para document.title.
 * Usado por DocumentTitle em App.jsx; evita conflito entre páginas e MainLayout.
 */
export const PAGE_TITLES = {
  '/login': 'Login',
  '/register': 'Criar conta',
  '/terms': 'Termos de Uso',
  '/support': 'Suporte',
  '/onboarding': 'Finalize seu cadastro',
  '/welcome': 'Fábrica conectada',
  '/': 'Dashboard',
  '/dashboard': 'Dashboard',
  '/assets': 'Gestão de Ativos',
  '/ia': 'NXD Intelligence',
  '/import': 'Importar histórico',
  '/financial': 'Indicadores financeiros',
  '/settings': 'Ajustes',
  '/billing': 'Cobrança',
};

const DEFAULT_TITLE = 'NXD';

/**
 * Rota desconhecida (ex.: 404) retorna "NXD"; NotFound seta "Página não encontrada | NXD" no mount.
 * Assim não há flicker entre DocumentTitle e NotFound.
 */
export function getPageTitle(pathname) {
  const title = PAGE_TITLES[pathname];
  if (title) return `${title} | NXD`;
  return DEFAULT_TITLE;
}
