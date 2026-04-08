import { test, expect } from '@playwright/test';

test('user can sign up successfully and log in', async ({ page }) => {
  // Part 1 - register user
  await page.goto('http://localhost:8080/register-user');

  // User data
  const timestamp = Date.now();
  const username = `pony${timestamp}`;
  const email = `pony${timestamp}@example.com`;
  const password = 'password';

  // Fill out form
  await page.fill('input[name="username"]', username);
  await page.fill('input[name="email"]', email);
  await page.fill('input[name="password"]', password);
  await page.fill('input[name="password2"]', password);

  // Submit form
  await page.click('input[value="Sign Up"]');

  // Check if an error is displayed
const error = page.locator('div.error');
if (await error.isVisible()) {
  console.log('Signup error:', await error.textContent());
}

// Part 2 - Log in with new user
await page.goto('http://localhost:8080/login');

await page.fill('input[name="username"]', username);
await page.fill('input[name="password"]', password);
await page.click('input[value="Sign In"]');

const loginError = page.locator('div.error');
await expect(loginError).toHaveCount(0);

// After log in we excpect to be on the user timeline
await expect(page).toHaveURL('http://localhost:8080');
await expect(page.locator('h2')).toHaveText('My Timeline');
await expect(page.locator('div.twitbox')).toContainText(`What's on your mind ${username}?`);
const messages = page.locator('ul.messages');
await expect(messages).toBeVisible();
})
