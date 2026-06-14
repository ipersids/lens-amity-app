import { type SubmitEventHandler, useState } from "react";
import { useNavigate } from "react-router";
import { useLoading, useSignup } from "../../../stores/auth";
import PasswordField from "./PasswordField";

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

  return (
    <form className="auth-form" onSubmit={handleSubmit}>
      <label>
        Username
        <input
          autoComplete="username"
          disabled={isLoading}
          name="username"
          onChange={(event) => setUsername(event.target.value)}
          required
          type="text"
          value={username}
        />
      </label>

      <PasswordField
        autoComplete="new-password"
        name="password"
        onChange={(event) => setPassword(event.target.value)}
        value={password}
        disabled={isLoading}
        label="New password"
      />

      {error && <p className="auth-error">{error}</p>}

      <button disabled={isLoading} type="submit">
        {"Sign Un"}
      </button>
    </form>
  );
};

export default SignupForm;
