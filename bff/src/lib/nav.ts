import * as m from '#lib/paraglide/messages.js';

export type Role = 'advisor' | 'client';

export interface NavItem {
	href: string;
	icon: string;
	label: () => string;
	/** Number of items that need the user's attention; omitted or 0 shows no badge. */
	count?: number;
}

/** Rail layout: `top` items sit at the top, `bottom` items are pinned to the bottom. */
export interface NavStructure {
	top: NavItem[];
	bottom: NavItem[];
}

export const settingsItem: NavItem = {
	href: '/settings',
	icon: 'settings',
	label: () => m.nav_settings()
};

export const advisorNav: NavStructure = {
	top: [
		{ href: '/', icon: 'dashboard', label: () => m.nav_overview() },
	],
	bottom: [settingsItem]
};

export const clientNav: NavStructure = {
	top: [
		{ href: '/', icon: 'home', label: () => m.nav_start() },
	],
	bottom: [settingsItem]
};

export function navFor(role: Role): NavStructure {
	return role === 'advisor' ? advisorNav : clientNav;
}

export function isActive(href: string, pathname: string): boolean {
	return href === '/' ? pathname === '/' : pathname === href || pathname.startsWith(`${href}/`);
}
