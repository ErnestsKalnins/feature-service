import {Component, OnInit} from '@angular/core';
import {Location} from "@angular/common";
import {Feature} from "../services/feature";
import {FeatureService} from "../services/feature.service";

@Component({
  selector: 'app-new-feature',
  templateUrl: './new-feature.component.html',
})
export class NewFeatureComponent implements OnInit {
  feature: Feature = {
    id: null,
    technicalName: '',
    displayName: null,
    description: null,
    expiresOn: null,
    inverted: false,
  };

  constructor(
    private featureService: FeatureService,
    private location: Location
  ) {
  }

  ngOnInit(): void {
  }

  goBack(): void {
    this.location.back();
  }

  invert(): void {
    this.feature.inverted = !this.feature.inverted;
  }

  saveFeature(): void {
    console.log(this.feature);
    this.featureService.saveFeature(this.feature)
      .subscribe();
    // TODO: redirect to list view.
  }
}
