import { useMutation, useQueryClient } from "@tanstack/react-query";
import { ProfileQueryKey } from "../constants";
import avatarService from "../services/avatar";
import type { UserProfileResponse } from "../services/users";
import { useUser } from "../stores/auth";

const useDeleteAvatar = () => {
  const currentUser = useUser();
  const queryClient = useQueryClient();
  const profileKey = ProfileQueryKey(currentUser?.username ?? "");

  return useMutation({
    mutationFn: avatarService.deleteAvatar,
    onMutate: async () => {
      await queryClient.cancelQueries({ queryKey: profileKey, exact: true });

      const oldProfile = queryClient.getQueryData<UserProfileResponse>(profileKey);

      queryClient.setQueryData<UserProfileResponse>(profileKey, (profile) =>
        profile ? { ...profile, avatar: undefined } : profile,
      );

      return { oldProfile };
    },

    onError: (_error, _variables, onMutateResult) => {
      if (onMutateResult?.oldProfile) {
        queryClient.setQueryData(profileKey, onMutateResult.oldProfile);
      }
    },
  });
};

export default useDeleteAvatar;
