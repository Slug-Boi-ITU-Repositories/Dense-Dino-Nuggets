import { test, expect } from '@playwright/test';

test('public timeline loads', async ({ page }) => {
  await page.goto('http://localhost:8080/public');

  // Check the heading exists
  await expect(page.getByRole('heading', { level: 2 })).toBeVisible();

  // Messages list should exist
  const list = page.locator('ul.messages');
  await expect(list).toBeVisible();

  // At least one <li> exists (tweet OR empty state)
  await expect(list.locator('li').first()).toBeVisible();
});
