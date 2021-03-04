<template>
  <fragment>
    <v-container>
      <v-row
        align="center"
        justify="center"
        class="my-4"
      >
        <v-col
          sm="8"
        >
          <v-row>
            <v-col>
              <h3>
                Number of devices has your namespace
              </h3>
            </v-col>
          </v-row>

          <v-spacer />

          <div class="mt-4 mx-4">
            {{ `Device limit:
              ${countDevicesHasNamespace()}/${countDevicesHasNamespacePercent().maxDevices}` }}

            <v-progress-linear
              class="mt-2"
              :value="countDevicesHasNamespacePercent().percent"
            />
          </div>
        </v-col>
      </v-row>
    </v-container>
  </fragment>
</template>

<script>

export default {
  name: 'SettingBilling',

  async created() {
    try {
      await this.$store.dispatch('namespaces/get', localStorage.getItem('tenant'));
    } catch {
      this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.namespaceLoad);
    }
  },

  methods: {
    countDevicesHasNamespace() {
      return this.$store.getters['namespaces/get'].devices_count;
    },

    countDevicesHasNamespacePercent() {
      const maxDevices = this.$store.getters['namespaces/get'].max_devices;

      let percent = 0;
      if (maxDevices >= 0) {
        percent = (this.countDevicesHasNamespace() / maxDevices) * 100;
        return { maxDevices, percent };
      }
      return { maxDevices, percent };
    },
  },
};
</script>
