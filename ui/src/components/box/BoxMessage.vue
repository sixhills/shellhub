<template>
  <fragment>
    <v-card
      class="mx-auto py-6"
      outlined
    >
      <!-- max-width="600" -->
      <v-card-title class="justify-center">
        <div>
          <v-icon
            x-large
          >
            security
          </v-icon>
        </div>
      </v-card-title>

      <v-card-title class="justify-center">
        <div class="headline">
          Looks like you don't have any Firewall Rule
        </div>
      </v-card-title>

      <v-list-item class="text-center">
        <v-list-item-content class="justify-center">
          <p>
            ShellHub provides flexible firewall for filtering SSH connections.
            It gives a fine-grained control over which SSH connections reach the devices.
          </p>
          <p class="mt-3">
            Using Firewall Rules you can deny or allow SSH connections from specific
            IP address to a specific or a group of devices using a given username.
          </p>
        </v-list-item-content>
      </v-list-item>

      <v-card-text class="pt-0 pb-0 text-center">
        <p>
          Firewall rules gives a fine-grained control over
          which SSH connections reach the devices.
          <a
            target="_blank"
            href="https://docs.shellhub.io/user-manual/managing-firewall-rules/"
          >See More</a>
        </p>
      </v-card-text>

      <v-card-actions class="justify-center pt-0">
        <FirewallRuleFormDialog
          :create-rule="true"
        />
      </v-card-actions>
    </v-card>
  </fragment>
</template>

<script>

import FirewallRuleFormDialog from '@/components/firewall_rules/FirewallRulesFormDialog';

export default {
  name: 'BoxMessage',

  components: {
    FirewallRuleFormDialog,
  },

};
</script>
