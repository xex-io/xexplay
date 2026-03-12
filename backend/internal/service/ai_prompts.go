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
6. Provide translations in English (en), Persian/Farsi (fa), and Arabic (ar)
7. Questions should be engaging, varied, and relevant to the specific match

Return ONLY a JSON array with this exact structure (no other text):
[
  {
    "question_text": {"en": "...", "fa": "...", "ar": "..."},
    "tier": "gold|silver|white",
    "high_answer_is_yes": true|false|null,
    "resolution_criteria": "Clear description of how to resolve this question"
  }
]`

const eventContentPrompt = `Generate a short event name and description for a sports league event in a prediction game app.

Sport/League: %s (%s)

Return ONLY a JSON object with this structure (no other text):
{
  "name": {"en": "...", "fa": "...", "ar": "..."},
  "description": {"en": "...", "fa": "...", "ar": "..."}
}`

const autoResolvePrompt = `You are a sports match result analyzer. Given a prediction question and the actual match result, determine if the answer is YES or NO.

Match: %s vs %s
Final Score: %s %d - %d %s

Question: %s
Resolution Criteria: %s

Based on the match result and the resolution criteria, is the answer YES or NO?
Return ONLY a JSON object: {"answer": true} for YES or {"answer": false} for NO.`
