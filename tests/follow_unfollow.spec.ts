import { test, expect } from '@playwright/test';

test('user follow and unfollow testuser', async ({ page }) => {
  // Part 1 - register user since we do not have any user to log in with
  await page.goto('http://server:8080/register-user');

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
  await page.goto('http://server:8080/login');

  await page.fill('input[name="username"]', username);
  await page.fill('input[name="password"]', password);
  await page.click('input[value="Sign In"]');

  const loginError = page.locator('div.error');
  await expect(loginError).toHaveCount(0);

  // Part 3 - Follow and Unfollow testuser
  // Follow testuser
  await page.goto('http://server:8080/testuser');

  const followStatus1 = page.locator('div.followstatus');
  await expect(followStatus1).toContainText('You are not yet following this user.');

  const followLink = page.locator('a.follow');
  await expect(followLink).toBeVisible();
  await followLink.click();

  const followStatus2 = page.locator('div.followstatus');
  await expect(followStatus2).toContainText('You are currently following this user.');

  // TODO Add test that user shows up in my "timeline"


  // Unfollow testuser
  await page.goto('http://server:8080/testuser');

  const unfollowLink = page.locator('a.unfollow');
  await expect(unfollowLink).toBeVisible();
  await unfollowLink.click();

  const followStatus3 = page.locator('div.followstatus');
  await expect(followStatus3).toContainText('You are not yet following this user.');
})