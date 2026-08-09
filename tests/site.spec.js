import { test, expect } from '@playwright/test'

const externalDestinations = {
  meetup: 'https://www.meetup.com/prague-golang-meetup/',
  meetupEvent: 'https://www.meetup.com/prague-golang-meetup/events/315974056/',
  slack: 'https://invite.slack.golangbridge.org/',
  mastodon: 'https://fosstodon.org/@gomeetupprague',
  youtube: 'https://www.youtube.com/@gomeetupprague',
  anniversary: 'https://10years.gomeetupprague.cz/',
}

test.beforeEach(async ({ page }) => {
  // The application is static. Blocking third-party requests keeps the suite
  // deterministic while still allowing every external destination to be checked.
  await page.route(/^https:\/\//, (route) => route.abort())
})

test('homepage exposes every primary community action', async ({ page }) => {
  const response = await page.goto('/')

  expect(response?.ok()).toBeTruthy()
  await expect(page).toHaveTitle('Go Meetup Prague')
  await expect(page.getByRole('heading', { level: 1 })).toContainText(
    'Build better software.',
  )
  await expect(page.locator('.meetup-alert')).toHaveAttribute(
    'href',
    externalDestinations.meetupEvent,
  )
  await expect(page.locator('.meetup-alert-details')).toBeVisible()
  await expect(page.locator('.meetup-alert-details')).toContainText(
    'The next Prague Go meetup is coming on Sep 23 at 6:00 PM at Sky Czechia Afi Karlin.',
  )
  await expect(page.getByRole('link', { name: /join the next meetup/i })).toHaveAttribute(
    'href',
    externalDestinations.meetupEvent,
  )
  await expect(page.getByRole('link', { name: /meet us in/i })).toHaveAttribute(
    'href',
    externalDestinations.slack,
  )
  await expect(page.getByRole('heading', { name: 'Take the stage.' })).toBeVisible()

  const footer = page.getByRole('contentinfo')
  await expect(footer.getByRole('link', { name: 'Mastodon' })).toHaveAttribute(
    'href',
    externalDestinations.mastodon,
  )
  await expect(footer.getByRole('link', { name: 'YouTube' })).toHaveAttribute(
    'href',
    externalDestinations.youtube,
  )
  await expect(footer.getByRole('link', { name: '10 years' })).toHaveAttribute(
    'href',
    externalDestinations.anniversary,
  )
})

test('past meetup banner is hidden', async ({ page }) => {
  await page.clock.setFixedTime(new Date('2026-09-23T19:00:01Z'))
  await page.goto('/')

  await expect(page.locator('.meetup-alert')).toBeHidden()
})

test('navigation opens the complete video archive', async ({ page }) => {
  await page.goto('/')
  await page.getByRole('link', { name: /watch past talks/i }).click()

  await expect(page).toHaveURL(/\/videos\/$/)
  await expect(page.getByRole('heading', { level: 1, name: 'Videos' })).toBeVisible()

  const cards = page.locator('.video-card')
  await expect(cards).toHaveCount(42)
  await expect(cards.first().getByRole('heading', { level: 3 })).not.toBeEmpty()

  const embeds = page.locator('.video-embed iframe')
  await expect(embeds).toHaveCount(42)
  expect(await embeds.evaluateAll((frames) => frames.every((frame) => frame.loading === 'lazy'))).toBe(
    true,
  )
  expect(
    await embeds.evaluateAll((frames) =>
      frames.every((frame) => /^https:\/\/(www\.)?youtube\.com\/embed\//.test(frame.src)),
    ),
  ).toBe(true)
})

test('layout fits the active desktop or phone viewport', async ({ page, isMobile }) => {
  await page.goto('/')

  const viewport = page.viewportSize()
  expect(viewport).not.toBeNull()
  const documentWidth = await page.evaluate(() => document.documentElement.scrollWidth)
  expect(documentWidth).toBeLessThanOrEqual(viewport.width + 1)

  await expect(page.locator('.hero')).toBeVisible()
  await expect(page.locator('.community-section')).toBeVisible()
  await expect(page.locator('.speaker-callout')).toBeVisible()

  const homeNavigation = page.getByRole('navigation').getByRole('link', { name: 'Home' })
  if (isMobile) {
    await expect(homeNavigation).toBeHidden()
    await expect(page.getByRole('link', { name: /join the community/i })).toBeVisible()
  } else {
    await expect(homeNavigation).toBeVisible()
  }
})

test('keyboard users can skip directly to the main content', async ({ page }) => {
  await page.goto('/')
  const skipLink = page.getByRole('link', { name: 'Skip to content' })
  await skipLink.focus()
  await expect(skipLink).toBeFocused()
  await skipLink.press('Enter')
  await expect(page.locator('#main-content')).toBeFocused()
})
