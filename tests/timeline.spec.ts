import { test, expect } from '@playwright/test';

test('public timeline loads', async ({ page }) => {
  await page.goto('http://localhost:8080/public');

  // Check the heading exists
  await expect(page.getByRole('heading', { level: 2 })).toBeVisible();

  // Messages list should exist
  const list = page.locator('ul.messages');
  await expect(list).toBeVisible();

  // Get first tweet
  const firstTweet = list.locator('li').first()
  await expect(firstTweet).toBeVisible();

  // Tweet renders properly
  // Author/User
  const authorLink = firstTweet.locator('a');
  await expect(authorLink).toBeVisible();
  await expect(authorLink).not.toHaveText(''); // Username not empty

  // Text - We have to trim the username and timestamp to check text content
  const tweetText = await firstTweet.locator('p').evaluate((p) => {
  const strong = p.querySelector('strong');
  const small = p.querySelector('small');
  let text = p.textContent || '';
  if (strong) text = text.replace(strong.textContent || '', '');
  if (small) text = text.replace(small.textContent || '', '');
  return text.trim();
  });

  expect(tweetText.length).toBeGreaterThan(0);
  

  // Timestamp
  const timestamp = firstTweet.locator('small');
  await expect(timestamp).toBeVisible();
  await expect(timestamp).not.toHaveText('');

});
