# Accessibility

This page says what hourdial does for people who use a screen reader, the
keyboard or reduced motion, and how far that has been checked. If something
here does not match what you hear or see, please open an issue.

## How this was checked

The menu bar app and the notch were checked by reading their code and the
accessibility attributes it sets (labels, roles, values, actions), and the
words VoiceOver is given are pinned by tests. **Nobody has yet used hourdial end
to end with VoiceOver running and listened to it.** Until that happens, treat
what follows as what the code is built to do, not as a tested experience.

## Menu bar item

- The status item has a spoken label that says the state, not the glyph:
  "hourdial, 2 hours 54 minutes tracked today", "hourdial, TD 3, 19 minutes left",
  "hourdial, 1 hour tracked today, trackd is not running", "hourdial, not set up
  yet". The label changes once a minute during a session, not every second.
- The menu is a standard macOS menu, so VoiceOver reads it natively. Rows
  whose titles use short forms carry their own spoken label: "Rate ★4" is
  "Rate 4 stars", "Inbox · 48m" is "Inbox, 48 minutes", a countdown such as
  "19:00 left" is "19 minutes left".
- The colour strip of the day, which is a row of coloured blocks, is read
  as shares instead: "Today: deep 33%, shallow 25%, distracting 8%, no label
  33%".
- In the start prompt (⌥⌘F) the goal field and the length menu are named.
- The break screen's progress bar is hidden from VoiceOver; the clock above
  it says how long the break has run, and its buttons are standard buttons.

## Notch panel

The notch draws most of what it shows itself, which VoiceOver cannot see
unless the code describes it. Every drawn part now does:

- The notch is a group named "hourdial notch". VoiceOver's press opens or closes
  the panel; Escape (or VoiceOver's cancel) closes it.
- The face either side of the camera reads as one line ("Focus session: TD
  3", "Standup, in 4 minutes", "hourdial, 85% of its budget"), with a running
  countdown given as its value ("19 minutes left").
- Every panel button is a button with a press action that does what a click
  does. Start buttons say what they do ("Start a 25 minute focus session",
  not "50m"). Label chips in the inbox round say the label and, for a guess,
  how likely it is; their number key is given as help, and the picked one is
  marked selected.
- Number rows, bar charts, progress rings, the 24-hour day dial, the
  project rows and the colour strips each expose one element per figure with
  words written for speech, for example "Thu, today: 5 hours; deep 3 hours,
  distracting 15 minutes; score 79", "Study: 1 hour of 2 hours, short, 3 day
  streak", or "The day as a dial: deep 09:00 to 11:00, away 11:00 to
  12:00; now 17:20; 5 hours tracked".
- Each page is a group named after it. Only the page in view is exposed, not
  the ones clipped beside it. The page dots are a stepper ("Page, 2 of 4,
  This week"): increment and decrement turn the page, as the arrow keys and
  a two-finger swipe do.
- Decoration (kind dots, the app letter badge) is hidden.

## Known gaps

- The notch only opens from the pointer, a click, or VoiceOver's press once
  VoiceOver has reached it. It sits in a panel above the menu bar, and how
  easily VoiceOver navigation reaches that panel has not been tried.
- Resizing the panel is pointer only (corner grips or a pinch).
- Moments the notch opens by itself (a session ending, the recap) are not
  announced; they also arrive as notifications when notifications are on.
- The layout code (widths, wrapping, what is cut for room) is
  AppKit drawing and has no tests; the words and their order are tested in
  `internal/bar` (`golden_test.go`, `spoken_test.go`).

## Reduced motion

With Reduce motion on in System Settings, the notch and the break screen
jump or cross-fade instead of springing and sliding.
