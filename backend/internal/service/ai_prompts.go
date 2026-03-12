package service

const cardGenerationPrompt = `You are a sports prediction question generator for a mobile prediction game app called XEX Play.

Given the following match details:
- Sport: %s (%s)
- Match: %s vs %s
- Kickoff: %s
- Odds (Home/Away/Draw): %.2f / %.2f / %.2f

Generate exactly %d yes/no prediction questions with this tier distribution:
- Gold (hardest, unlikely/surprising outcomes): %d questions
- Silver (moderate difficulty, plausible but uncertain): %d questions
- White (easiest, likely or common outcomes): %d questions

IMPORTANT RULES:
1. Each question must be answerable with YES or NO after the match concludes
2. Include a clear "resolution_criteria" explaining exactly how to determine the answer
3. For Gold questions: the "high_answer_is_yes" should be true if YES is the unlikely/high-reward answer
4. For Silver questions: the "high_answer_is_yes" should be true if YES is the slightly less likely answer
5. For White questions: set "high_answer_is_yes" to null (these have flat scoring)
6. Provide translations in ALL of these languages:
   - en: English
   - fa: Persian (Farsi)
   - ar: Arabic
   - tr: Turkish
   - es: Spanish
   - ru: Russian
   - zh: Chinese (Simplified)
   - id: Indonesian (Bahasa Indonesia)
   - vi: Vietnamese
   - ms: Malay (Bahasa Melayu)
7. Questions should be engaging, varied, and relevant to the specific match
8. Use natural, native-sounding phrasing in each language — not word-for-word translation

Return ONLY a JSON array with this exact structure (no other text):
[
  {
    "question_text": {"en": "...", "fa": "...", "ar": "...", "tr": "...", "es": "...", "ru": "...", "zh": "...", "id": "...", "vi": "...", "ms": "..."},
    "tier": "gold|silver|white",
    "high_answer_is_yes": true|false|null,
    "resolution_criteria": "Clear description of how to resolve this question"
  }
]`

const eventContentPrompt = `Generate a short event name and description for a sports league event in a prediction game app.

Sport/League: %s (%s)

Provide translations in ALL of these languages: en (English), fa (Persian), ar (Arabic), tr (Turkish), es (Spanish), ru (Russian), zh (Chinese Simplified), id (Indonesian), vi (Vietnamese), ms (Malay).
Use natural, native-sounding phrasing in each language.

Return ONLY a JSON object with this structure (no other text):
{
  "name": {"en": "...", "fa": "...", "ar": "...", "tr": "...", "es": "...", "ru": "...", "zh": "...", "id": "...", "vi": "...", "ms": "..."},
  "description": {"en": "...", "fa": "...", "ar": "...", "tr": "...", "es": "...", "ru": "...", "zh": "...", "id": "...", "vi": "...", "ms": "..."}
}`

const autoResolvePrompt = `You are a sports match result analyzer. Given a prediction question and the actual match result, determine if the answer is YES or NO.

Match: %s vs %s
Final Score: %s %d - %d %s

Question: %s
Resolution Criteria: %s

Based on the match result and the resolution criteria, is the answer YES or NO?
Return ONLY a JSON object: {"answer": true} for YES or {"answer": false} for NO.`

const teamNameTranslationPrompt = `Translate the following sports team names into all listed languages. Use the official/commonly known name in each locale when one exists (e.g. "Real Madrid" stays "Real Madrid" in most languages, but "Bayern Munich" becomes "بایرن مونیخ" in Persian). If no well-known local name exists, transliterate into the target script.

Teams to translate:
%s

Languages: en (English), fa (Persian), ar (Arabic), tr (Turkish), es (Spanish), ru (Russian), zh (Chinese Simplified), id (Indonesian), vi (Vietnamese), ms (Malay).

Return ONLY a JSON object mapping each original team name to its translations (no other text):
{
  "Team Name": {"en": "...", "fa": "...", "ar": "...", "tr": "...", "es": "...", "ru": "...", "zh": "...", "id": "...", "vi": "...", "ms": "..."}
}`
