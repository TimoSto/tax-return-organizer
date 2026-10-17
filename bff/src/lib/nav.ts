import * as m from '#lib/paraglide/messages.js';

export type Role = 'advisor' | 'client';

export interface NavItem {
	href: string;
	icon: string;
	label: () => string;
	/** Number of items that need the user's attention; omitted or 0 shows no badge. */
	count?: number;
}

export interface NavGroup {
	heading: () => string;
	items: NavItem[];
}

export const settingsItem: NavItem = {
	href: '/einstellungen',
	icon: 'settings',
	label: () => m.nav_settings()
};

export const advisorNav: NavGroup[] = [
	{
		heading: () => m.nav_group_firm(),
		items: [
			{ href: '/', icon: 'dashboard', label: () => m.nav_overview() },
			{ href: '/mandanten', icon: 'group', label: () => m.nav_clients() },
			{ href: '/personas', icon: 'badge', label: () => m.nav_personas() },
			{ href: '/fristen', icon: 'event', label: () => m.nav_deadlines() }
		]
	},
	{
		heading: () => m.nav_group_collaboration(),
		items: [{ href: '/nachrichten', icon: 'chat_bubble', label: () => m.nav_messages() }]
	}
];

export const clientNav: NavGroup[] = [
	{
		heading: () => m.nav_group_firm(),
		items: [
			{ href: '/', icon: 'home', label: () => m.nav_start() },
			{ href: '/aufgaben', icon: 'checklist', label: () => m.nav_tasks() },
			{ href: '/unterlagen', icon: 'folder', label: () => m.nav_documents() },
			{ href: '/nachrichten', icon: 'chat_bubble', label: () => m.nav_messages() }
		]
	}
];

export function navFor(role: Role): NavGroup[] {
	return role === 'advisor' ? advisorNav : clientNav;
}

export function isActive(href: string, pathname: string): boolean {
	return href === '/' ? pathname === '/' : pathname === href || pathname.startsWith(`${href}/`);
}
