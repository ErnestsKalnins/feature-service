import {Component, OnInit} from '@angular/core';
import {Feature} from "../services/feature";
import {FeatureService} from "../services/feature.service";

@Component({
  selector: 'feature-list',
  templateUrl: './feature-list.component.html',
})
export class FeatureListComponent implements OnInit {
  features: Feature[] = [];

  constructor(
    private featureService: FeatureService,
  ) {
  }

  ngOnInit(): void {
    this.getFeatures();
  }

  getFeatures(): void {
    this.featureService.getFeatures()
      .subscribe(features => this.features = features);
  }
}
