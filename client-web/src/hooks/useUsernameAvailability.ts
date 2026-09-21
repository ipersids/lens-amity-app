import { useCallback, useEffect, useState } from "react";
import { validateUsername } from "../features/auth/validation";
import { getApiError } from "../services/api";
import authService from "../services/auth";

type UsernameAvailabilityStatus =
  | "default"
  | "invalid"
  | "checking"
  | "available"
  | "unavailable"
  | "error"
  | "unchanged";

const usernameStatusMessages: Record<UsernameAvailabilityStatus, string> = {
  default: "Usernames are first-come, first-served.",
  invalid: "Username is invalid. Try again.",
  checking: "Checking availability...",
  available: "This username is available.",
  unavailable: "This username is already taken. Try another.",
  error: "We could not check availability. Try again.",
  unchanged: "This is already your username.",
};

const useUsernameAvailability = ({
  newUsername,
  currentUsername,
  delay,
}: {
  newUsername: string;
  currentUsername: string;
  delay: number;
}) => {
  const [status, setStatus] = useState<UsernameAvailabilityStatus>("default");
  const candidate = newUsername.trim();
  const validationError = validateUsername(candidate);
  const [statusMsg, setStatusMsg] = useState<string>(validationError);

  const handleStatus = useCallback((nextStatus: UsernameAvailabilityStatus, reason?: string) => {
    setStatus(nextStatus);
    setStatusMsg(reason ?? usernameStatusMessages[nextStatus]);
  }, []);

  useEffect(() => {
    if (!candidate) {
      handleStatus("default");
      return;
    }

    if (validationError) {
      handleStatus("invalid", validationError);
      return;
    }

    if (candidate.toLocaleLowerCase() === currentUsername) {
      handleStatus("unchanged");
      return;
    }

    handleStatus("checking");

    const controller = new AbortController();
    const timeoutID = window.setTimeout(async () => {
      try {
        const { isAvailable, validationError } = await authService.getUsernameAvailability(
          candidate,
          controller.signal,
        );
        if (!controller.signal.aborted) {
          handleStatus(isAvailable ? "available" : "unavailable", validationError);
        }
      } catch (err: unknown) {
        if (!controller.signal.aborted) {
          const msg = getApiError(err);
          handleStatus("error", msg.error.message);
        }
      }
    }, delay);

    return () => {
      window.clearTimeout(timeoutID);
      controller.abort();
    };
  }, [handleStatus, candidate, delay, currentUsername, validationError]);

  return { status, statusMsg };
};

export default useUsernameAvailability;
