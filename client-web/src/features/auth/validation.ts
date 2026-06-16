import zxcvbn from "zxcvbn";

type PaswordValidationResult = {
  score: number;
  description: string;
  feedback: string;
};

const USERNAME_REGEX = /^[A-Za-z0-9_-]+$/;
const MIN_PASSWORD_LENGTH = 15;

const scoreToWord = (score: number): string => {
  return ["Very Weak", "Very Weak", "Weak", "Good", "Strong"][score] ?? "Unknown";
};

export const validatePassword = (
  password: string,
  input: string[] = [],
): PaswordValidationResult => {
  const res = zxcvbn(password, input);

  if (res.score > 2 && password.length < MIN_PASSWORD_LENGTH) {
    return {
      score: 2,
      description: "Good",
      feedback: "Just a few more characters to reach the recommended minimum of 15",
    };
  }

  return {
    score: res.score,
    description: scoreToWord(res.score),
    feedback: res.feedback.warning || res.feedback.suggestions[0] || "",
  };
};

export const validateUsername = (username: string): string => {
  return username && !USERNAME_REGEX.test(username)
    ? "Use Latin letters, numbers, hyphens, and underscores"
    : "";
};
