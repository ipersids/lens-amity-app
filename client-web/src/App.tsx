import { useEffect } from "react";
import { Route, Routes } from "react-router";
import Layout from "./features/Layout";
import AuthPage from "./pages/AuthPage";
import { useSyncSession } from "./stores/auth";

function App() {
  const syncSession = useSyncSession();

  useEffect(() => {
    syncSession();
  }, [syncSession]);

  return (
    <div className="app">
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<p>APP</p>} />
          <Route path="/login" element={<AuthPage mode="login" />} />
          <Route path="/signup" element={<AuthPage mode="signup" />} />
        </Route>
      </Routes>
    </div>
  );
}

export default App;
