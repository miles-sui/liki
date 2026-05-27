# Frequently Asked Questions

## How does the assessment work?

The assessment has 6 rounds. Each round shows three short statements about behavior, preference, or attitude. You pick the two that are most like you and rank them first and second. The unchosen statement is treated as least applicable.

There are no right or wrong answers. The questions are designed so that every possible response pattern maps to a valid position in the Five Elements space. It takes a few minutes.

Your result is computed instantly after the final round. If you are logged in, your assessment is saved. If not, you see your result immediately and can register to keep it.

## How accurate is my result?

Your result reflects the choices you made at this moment. It captures tendencies — not permanent traits. Think of it as a snapshot, not a destiny.

If you retake the assessment weeks or months later, your result may shift slightly. That is normal and informative — it shows how your self-perception evolves. If your result shifts dramatically, it likely reflects a genuine change in how you see yourself or a different mindset when answering.

This is not a measurement of a fixed quantity. It is a reflection of your own expressed preferences. The tool reflects what you tell it — if you answer honestly, the result faithfully captures your self-perception at that moment.

## Do I need to provide my birth date?

No. The assessment is based entirely on your own choices — the ranked-choice questions. Birth date, time, or location are not required and are not collected.

The Flow calendar uses solar terms — 24 astronomical markers based on the sun's position — as a seasonal overlay for your profile. This is an astronomical calendar, not an astrological one, and it does not require any birth data. It answers "how does your shape interact with the time of year," not "what were you born under."

## Can other people see my result?

Not by default. Your profile is private. Only you see your full assessment result.

You can optionally set your profile to public, which makes your identity label and element profile visible to other logged-in users. You can switch back to private at any time.

Peer reviewers see only the questions they are answering and their own summary — never your self-assessment result.

## What is the difference between pure types and composite types?

A pure type means one element is clearly dominant over the others — for example, Wood (Pioneer) or Water (Reservoir). There are 5 pure types.

A composite type means two elements are significantly elevated, with one leading. The boundary is defined by a 42-degree angle on the unit sphere — a mathematical choice that distributes the types approximately evenly across the five-element space. There are 20 composite types: 10 Sheng-pair (elements that nourish each other) and 10 Ke-pair (elements that restrain each other).

Most people are composites. That is expected. Being a pure type is not "purer" or "better" — it simply means your responses showed one element clearly ahead of the rest.

## What does Bond tell me?

Bond computes the directional forces between two personality profiles using the Sheng (nourishing) and Ke (restraining) matrices. The result is a two-way force map:

- What your dominant elements do to theirs (facilitation or restraint)
- What their dominant elements do to yours
- Whether the exchange is roughly symmetrical or strongly directional

Bond does not give you a percentage or a verdict. It describes the dynamic — who brings what to the interaction. The language is deliberately neutral: "facilitates," "restrains," "resonates with." No "good match" or "bad match."

One full Bond is free for your first match. Passport subscribers get unlimited Bonds.

## How do peer reviews work?

You create a review link from your assessment result page (or account settings) and share it with friends. Each friend follows the link and answers the same type of five-element questions — but about you, not themselves.

When at least 3 people have submitted reviews, you can view a peer profile: an aggregated picture of how others see you, displayed next to your self-assessment for comparison.

Individual reviewer answers are never visible to you or to other reviewers. The aggregation threshold of 3 is a privacy minimum — with fewer reviewers, individual answers might be inferable.

Review links expire after 30 days. You can create new ones at any time.

## How do I cancel my subscription?

You can cancel through your account settings or by contacting us. Cancellation stops future renewal charges. Your Passport features remain active until the end of the current billing period.

There is no long-term commitment. You can subscribe one month, cancel, and come back whenever you want — your data stays intact.

See our Refund Policy for details on whether the current period is refundable (generally, it is not — this is standard for digital services delivered immediately).

## What happens if I delete my account?

Account deletion has two stages:

**Immediate** — all your active sessions are invalidated. You are logged out everywhere. A 7-day grace period begins.

**Within 7 days** — if you log back in, your account is fully restored. Everything is as it was.

**After 7 days** — your account is anonymized. Your username is replaced with a generic label, your email and password hash are cleared, and your assessments are detached from your identity. Peer reviews, match requests, and subscriptions are cleaned up per our data retention policy.

Some anonymized assessment data may be retained for aggregate model statistics. You will not be identifiable from this data.

## Is my data secure?

We take a security-minimal approach: the less data we hold, the less there is to protect.

- **Passwords**: hashed with argon2id (47104 KiB memory, 1 iteration, 4 parallelism). Even if our database were compromised, your password would not be exposed.
- **Authentication**: JWT with server-side `token_version`. Changing your password or logging out instantly invalidates all sessions on all devices.
- **Encryption in transit**: all traffic is HTTPS via Caddy.
- **Rate limiting**: authentication endpoints are rate-limited to prevent brute-force attempts.
- **No third-party trackers**: no analytics scripts, no social media integrations that could leak data.
- **Minimal data**: we collect only what is necessary to provide the service.

Full details in our Privacy Policy and Cookie Policy.
