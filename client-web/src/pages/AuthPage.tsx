import { Navigate } from "react-router";
import { AuthShell, LoginForm, SignupForm } from "../features/auth";
import { useUser } from "../stores/auth";

type AuthPageProps = {
  mode: "signup" | "login";
};

const AuthPage = ({ mode }: AuthPageProps) => {
  const user = useUser();

  if (user) {
    return <Navigate to="/" replace />;
  }

  const isSignup = mode === "signup";
  const title = isSignup ? "Create account" : "Log in";
  const switchTo = isSignup ? "/login" : "/signup";
  const switchText = isSignup ? "Log in instead" : "Create an account";

  return (
    <AuthShell title={title} switchTo={switchTo} switchText={switchText}>
      {isSignup ? <SignupForm /> : <LoginForm />}
    </AuthShell>
  );
};

export default AuthPage;
