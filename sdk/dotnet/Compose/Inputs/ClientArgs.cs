// *** WARNING: this file was generated by pulumi. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

using System;
using System.Collections.Generic;
using System.Collections.Immutable;
using System.Threading.Tasks;
using Pulumi.Serialization;
using Pulumi;

namespace WTTech.Aem.Compose.Inputs
{

    public sealed class ClientArgs : global::Pulumi.ResourceArgs
    {
        /// <summary>
        /// Used when trying to connect to the AEM instance machine (often right after creating it). Need to be enough long because various types of connections (like AWS SSM or SSH) may need some time to boot up the agent.
        /// </summary>
        [Input("action_timeout")]
        public Input<string>? Action_timeout { get; set; }

        [Input("credentials")]
        private InputMap<string>? _credentials;

        /// <summary>
        /// Credentials for the connection type
        /// </summary>
        public InputMap<string> Credentials
        {
            get => _credentials ?? (_credentials = new InputMap<string>());
            set => _credentials = value;
        }

        [Input("settings", required: true)]
        private InputMap<string>? _settings;

        /// <summary>
        /// Settings for the connection type
        /// </summary>
        public InputMap<string> Settings
        {
            get => _settings ?? (_settings = new InputMap<string>());
            set => _settings = value;
        }

        /// <summary>
        /// Used when reading the AEM instance state when determining the plan.
        /// </summary>
        [Input("state_timeout")]
        public Input<string>? State_timeout { get; set; }

        /// <summary>
        /// Type of connection to use to connect to the machine on which AEM instance will be running.
        /// </summary>
        [Input("type", required: true)]
        public Input<string> Type { get; set; } = null!;

        public ClientArgs()
        {
        }
        public static new ClientArgs Empty => new ClientArgs();
    }
}
