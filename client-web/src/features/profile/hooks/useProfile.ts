import { useQuery } from "@tanstack/react-query";
import usersService from "../../../services/users";

export const ProfileQueryKey = (username: string) => ["profile", username];

const useProfile = (username?: string) =>
  useQuery({
    queryKey: ProfileQueryKey(username ?? ""),
    queryFn: () => usersService.getUserProfile(username ?? ""),
    enabled: !!username,
    staleTime: 60_000,
    gcTime: 10 * 60_000,
    retry: 1,
  });

export default useProfile;
