import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { UpdateProfileProps, UserProfileResponse } from "../../../services/users";
import usersService from "../../../services/users";
import { ProfileQueryKey } from "./useProfile";

const useUpdateProfile = (username: string) => {
  const queryClient = useQueryClient();
  const queryKey = ProfileQueryKey(username);

  return useMutation({
    mutationFn: usersService.updateUserProfile,
    onMutate: async (changes: UpdateProfileProps) => {
      await queryClient.cancelQueries({ queryKey });

      const previousProfile = queryClient.getQueryData<UserProfileResponse>(queryKey);

      queryClient.setQueryData<UserProfileResponse>(queryKey, (profile) =>
        profile
          ? {
              ...profile,
              ...changes,
              about: changes.about ?? "",
            }
          : profile,
      );

      return { previousProfile };
    },

    onError: (_error, _changes, context) => {
      if (context?.previousProfile) {
        queryClient.setQueryData(queryKey, context.previousProfile);
      }
    },

    onSettled: () => {
      queryClient.invalidateQueries({ queryKey });
    },
  });
};

export default useUpdateProfile;
