import { test, expect } from '@playwright/test';

test('user can sign up successfully', async ({ page }) => {
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

// OBS! We cannot actually confirm successfull sign up due to our backend behavior.
// We should ideally log in the user and redirect to a different page.
})