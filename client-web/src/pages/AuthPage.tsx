import { type SubmitEventHandler, useState } from "react";
import { Link, Navigate, useNavigate } from "react-router";
import { useLoading, useLogin, useSignup, useUser } from "../stores/auth";

type AuthPageProps = {
  mode: "signup" | "login";
};

const AuthPage = ({ mode }: AuthPageProps) => {
  const user = useUser();
  const signup = useSignup();
  const login = useLogin();
  const isLoading = useLoading();
  const navigate = useNavigate();

  const [username, setUsername] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);

  if (user) {
    return <Navigate to="/" replace />;
  }

  const isSignup = mode === "signup";
  const title = isSignup ? "Create account" : "Log in";
  const switchTo = isSignup ? "/login" : "/signup";
  const switchText = isSignup ? "Log in instead" : "Create an account";

  const handleSubmit: SubmitEventHandler<HTMLFormElement> = async (event) => {
    event.preventDefault();
    setError(null);

    const trimmedUsername = username.trim();
    const trimmedDisplayName = displayName.trim();

    if (!trimmedUsername || !password) {
      setError("Username and password are required.");
      return;
    }

    try {
      if (isSignup) {
        await signup({
          username: trimmedUsername,
          displayName: trimmedDisplayName || undefined,
          password,
        });
        navigate("/login", { replace: true });
        return;
      }

      await login({ username: trimmedUsername, password });
      navigate("/", { replace: true });
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Unable to continue.");
    }
  };

  return (
    <section className="auth-page">
      <form className="auth-form" onSubmit={handleSubmit}>
        <h1>{title}</h1>

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

        {isSignup && (
          <label>
            Display name
            <input
              autoComplete="name"
              disabled={isLoading}
              name="displayName"
              onChange={(event) => setDisplayName(event.target.value)}
              type="text"
              value={displayName}
            />
          </label>
        )}

        <label>
          Password
          <input
            autoComplete={isSignup ? "new-password" : "current-password"}
            disabled={isLoading}
            name="password"
            onChange={(event) => setPassword(event.target.value)}
            required
            type="password"
            value={password}
          />
        </label>

        {error && <p className="auth-error">{error}</p>}

        <button disabled={isLoading} type="submit">
          {isLoading ? "Please wait..." : title}
        </button>

        <Link to={switchTo}>{switchText}</Link>
      </form>
    </section>
  );
};

export default AuthPage;
