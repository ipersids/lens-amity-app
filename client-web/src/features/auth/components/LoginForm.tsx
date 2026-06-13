import { EyeIcon, EyeSlashIcon } from "@heroicons/react/24/outline";
import { type SubmitEventHandler, useState } from "react";
import { useNavigate } from "react-router";
import { useLoading, useLogin } from "../../../stores/auth";

const LoginForm = () => {
  const login = useLogin();
  const isLoading = useLoading();
  const navigate = useNavigate();

  const [username, setUsername] = useState<string>("");
  const [password, setPassword] = useState<string>("");
  const [error, setError] = useState<string | null>(null);
  const [showPassword, setShowPassword] = useState<boolean>(false);

  const handleSubmit: SubmitEventHandler<HTMLFormElement> = async (event) => {
    event.preventDefault();

    const trimmedUsername = username.trim();

    if (!trimmedUsername || !password) {
      setError("Username and password are required.");
      return;
    }

    try {
      await login({ username: trimmedUsername, password });
      navigate("/", { replace: true });
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Oops, something went wrong.");
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

      <label>
        Password
        <div className="password-field">
          <input
            autoComplete="current-password"
            disabled={isLoading}
            name="password"
            onChange={(event) => setPassword(event.target.value)}
            required
            type={showPassword ? "text" : "password"}
            value={password}
          />
          <button
            type="button"
            onClick={() => setShowPassword((current) => !current)}
            disabled={isLoading}
            aria-label={showPassword ? "Hide password" : "Show password"}
          >
            {showPassword ? <EyeSlashIcon /> : <EyeIcon />}
          </button>
        </div>
      </label>

      {error && <p className="auth-error">{error}</p>}

      <button disabled={isLoading} type="submit">
        {isLoading ? "Making things done" : "Log in"}
      </button>
    </form>
  );
};

export default LoginForm;
