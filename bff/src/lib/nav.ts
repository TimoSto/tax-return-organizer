import * as m from '#lib/paraglide/messages.js';

export interface NavItem {
	href: string;
	icon: string;
	label: () => string;
	/** Number of items that need the user's attention; omitted or 0 shows no badge. */
	count?: number;
	/** Active only on `href` itself, not below it. For an area's home, whose href prefixes every other one. */
	exact?: boolean;
}

/** Rail layout: `top` items sit at the top, `bottom` items are pinned to the bottom. */
export interface NavStructure {
	top: NavItem[];
	bottom: NavItem[];
}

export const taxPracticeNav: NavStructure = {
	top: [{ href: '/tax-practice', icon: 'dashboard', label: () => m.nav_overview(), exact: true }],
	bottom: [{ href: '/tax-practice/settings', icon: 'settings', label: () => m.nav_settings() }]
};

export const clientNav: NavStructure = {
	top: [{ href: '/client', icon: 'home', label: () => m.nav_start(), exact: true }],
	bottom: [{ href: '/client/settings', icon: 'settings', label: () => m.nav_settings() }]
};

export function isActive(item: NavItem, pathname: string): boolean {
	return item.exact
		? pathname === item.href
		: pathname === item.href || pathname.startsWith(`${item.href}/`);
}
