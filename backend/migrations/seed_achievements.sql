INSERT INTO achievements (key, name, description, icon, category, condition_type, condition_value)
VALUES
    ('first_prediction',
     '{"en": "First Prediction", "fa": "\u0627\u0648\u0644\u06cc\u0646 \u067e\u06cc\u0634\u200c\u0628\u06cc\u0646\u06cc"}'::jsonb,
     '{"en": "Complete your first session", "fa": "\u0627\u0648\u0644\u06cc\u0646 \u062c\u0644\u0633\u0647 \u062e\u0648\u062f \u0631\u0627 \u06a9\u0627\u0645\u0644 \u06a9\u0646\u06cc\u062f"}'::jsonb,
     'trophy', 'milestone', 'first_prediction', 1),

    ('perfect_day',
     '{"en": "Perfect Day", "fa": "\u0631\u0648\u0632 \u0628\u06cc\u200c\u0646\u0642\u0635"}'::jsonb,
     '{"en": "All answers correct in a day", "fa": "\u0647\u0645\u0647 \u067e\u0627\u0633\u062e\u200c\u0647\u0627 \u062f\u0631 \u06cc\u06a9 \u0631\u0648\u0632 \u0635\u062d\u06cc\u062d"}'::jsonb,
     'star', 'performance', 'perfect_day', 1),

    ('streak_7',
     '{"en": "Week Warrior", "fa": "\u062c\u0646\u06af\u062c\u0648\u06cc \u0647\u0641\u062a\u0647"}'::jsonb,
     '{"en": "7-day streak", "fa": "\u06f7 \u0631\u0648\u0632 \u0645\u062a\u0648\u0627\u0644\u06cc"}'::jsonb,
     'fire', 'streak', 'streak_7', 7),

    ('streak_10',
     '{"en": "Dedicated Player", "fa": "\u0628\u0627\u0632\u06cc\u06a9\u0646 \u0645\u062a\u0639\u0647\u062f"}'::jsonb,
     '{"en": "10-day streak", "fa": "\u06f1\u06f0 \u0631\u0648\u0632 \u0645\u062a\u0648\u0627\u0644\u06cc"}'::jsonb,
     'fire', 'streak', 'streak_10', 10),

    ('streak_30',
     '{"en": "Monthly Master", "fa": "\u0627\u0633\u062a\u0627\u062f \u0645\u0627\u0647\u0627\u0646\u0647"}'::jsonb,
     '{"en": "30-day streak", "fa": "\u06f3\u06f0 \u0631\u0648\u0632 \u0645\u062a\u0648\u0627\u0644\u06cc"}'::jsonb,
     'crown', 'streak', 'streak_30', 30),

    ('champion',
     '{"en": "Daily Champion", "fa": "\u0642\u0647\u0631\u0645\u0627\u0646 \u0631\u0648\u0632\u0627\u0646\u0647"}'::jsonb,
     '{"en": "Rank #1 on daily leaderboard", "fa": "\u0631\u062a\u0628\u0647 \u06f1 \u062f\u0631 \u062c\u062f\u0648\u0644 \u0631\u0648\u0632\u0627\u0646\u0647"}'::jsonb,
     'medal', 'leaderboard', 'champion', 1),

    ('tournament_mvp',
     '{"en": "Tournament MVP", "fa": "\u0628\u0647\u062a\u0631\u06cc\u0646 \u0628\u0627\u0632\u06cc\u06a9\u0646 \u062a\u0648\u0631\u0646\u0645\u0646\u062a"}'::jsonb,
     '{"en": "Rank #1 in tournament", "fa": "\u0631\u062a\u0628\u0647 \u06f1 \u062f\u0631 \u062a\u0648\u0631\u0646\u0645\u0646\u062a"}'::jsonb,
     'medal', 'leaderboard', 'tournament_mvp', 1),

    ('referral_5',
     '{"en": "Social Butterfly", "fa": "\u067e\u0631\u0648\u0627\u0646\u0647 \u0627\u062c\u062a\u0645\u0627\u0639\u06cc"}'::jsonb,
     '{"en": "Refer 5 friends", "fa": "\u06f5 \u062f\u0648\u0633\u062a \u0645\u0639\u0631\u0641\u06cc \u06a9\u0646\u06cc\u062f"}'::jsonb,
     'people', 'social', 'referrals_5', 5),

    ('referral_10',
     '{"en": "Ambassador", "fa": "\u0633\u0641\u06cc\u0631"}'::jsonb,
     '{"en": "Refer 10 friends", "fa": "\u06f1\u06f0 \u062f\u0648\u0633\u062a \u0645\u0639\u0631\u0641\u06cc \u06a9\u0646\u06cc\u062f"}'::jsonb,
     'shield', 'social', 'referrals_10', 10)
ON CONFLICT (key) DO NOTHING;
