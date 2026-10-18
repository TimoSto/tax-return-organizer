import { expect, test } from '@playwright/test';

test('shows the collector name', async ({ page }) => {
	await page.goto('/advisor');
	await expect(page.locator('h1')).toHaveText('E2E Collector');
});
