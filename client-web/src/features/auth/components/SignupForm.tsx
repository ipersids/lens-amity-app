import { type SubmitEventHandler, useState } from "react";
import { useNavigate } from "react-router";
import { useLoading, useSignup } from "../../../stores/auth";
import { validatePassword, validateUsername } from "../validation";
import PasswordField from "./PasswordField";
import TextField from "./TextField";

const SignupForm = () => {
  const signup = useSignup();
  const isLoading = useLoading();
  const navigate = useNavigate();

  const [username, setUsername] = useState<string>("");
  const [password, setPassword] = useState<string>("");
  const [error, setError] = useState<string | null>(null);

  const handleSubmit: SubmitEventHandler<HTMLFormElement> = async (event) => {
    event.preventDefault();

    const trimmedUsername = username.trim();

    if (!trimmedUsername || !password) {
      setError("Username and password are required.");
      return;
    }

    try {
      await signup({
        username: trimmedUsername,
        displayName: trimmedUsername,
        password,
      });
      navigate("/login", { replace: true });
      return;
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Unable to continue.");
    }
  };

  const passwordValidation = validatePassword(password, [username]);

  return (
    <form id="signup" name="signup" className="auth-form" onSubmit={handleSubmit}>
      <TextField
        label={"Username"}
        id="username"
        name="username"
        autoComplete="username"
        type="text"
        value={username}
        onChange={(event) => setUsername(event.target.value)}
        disabled={isLoading}
        required
        error={validateUsername(username)}
      />

      <PasswordField
        id="new-password"
        autoComplete="new-password"
        name="new-password"
        onChange={(event) => setPassword(event.target.value)}
        value={password}
        disabled={isLoading}
        label="New password"
        strength={password ? passwordValidation.description : undefined}
        error={password ? passwordValidation.feedback : undefined}
      />

      {error && <p className="auth-error">{error}</p>}

      <button disabled={isLoading} type="submit">
        {"Create account"}
      </button>
    </form>
  );
};

export default SignupForm;
